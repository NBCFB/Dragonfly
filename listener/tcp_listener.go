package listener

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	"context"
)

type Listener struct {
	Address     string `json:"address"`
	FD       	int    `json:"fd"`
	Filename 	string `json:"filename"`
}

func CreateListener(address string) (net.Listener, error) {
	// Create a TCP _listener, as http _listener is a TCP _listener
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return listener, nil
}

func ImportListener(address string) (net.Listener, error) {
	// Extract the encoded _listener metadata from the environment.
	listenerEnv := os.Getenv("LISTENER")
	if listenerEnv == "" {
		return nil, errors.New("Unable to find LISTENER environment variable")
	}

	// Parse the _listener metadata.
	var listener Listener
	err := json.Unmarshal([]byte(listenerEnv), &listener)
	if err != nil {
		return nil, err
	}
	if listener.Address != address {
		return nil, fmt.Errorf("Unable to find _listener for %v", address)
	}

	// The file has already been passed to this process, extract the file
	// descriptor and name from the metadata to rebuild/find the *os.File for
	// the _listener.
	listenerFile := os.NewFile(uintptr(listener.FD), listener.Filename)
	if listenerFile == nil {
		return nil, fmt.Errorf("Unable to create _listener file: %v", err)
	}
	defer listenerFile.Close()

	// Create a net.Listener from the *os.File.
	newListener, err := net.FileListener(listenerFile)
	if err != nil {
		return nil, err
	}

	return newListener, nil
}


func CreateOrImportListener(address string) (net.Listener, error) {
	// Try and import a _listener for addr. If it's found, use it.
	listener, err := ImportListener(address)
	if err == nil {
		log.Printf("Imported _listener file descriptor for %v.\n", address)
		return listener, nil
	}

	// Since no _listener to be imported, create a _listener at the first
	listener, err = CreateListener(address)
	if err != nil {
		return nil, err
	}

	log.Printf("Created _listener file descriptor for %v.\n", address)
	return listener, nil
}

func GetListenerFile(listener net.Listener) (*os.File, error) {
	switch listenerType := listener.(type) {
	case *net.TCPListener:
		return listenerType.File()
	case *net.UnixListener:
		return listenerType.File()
	}
	return nil, fmt.Errorf("unsupported _listener: %T", listener)
}


func ForkChild(address string, listener net.Listener) (*os.Process, error) {
	// Parse the _listener metadata and pass it to the child in the environment.
	listenerFile, err := GetListenerFile(listener)

	log.Printf("Get _listener file at %v", listenerFile.Name())
	if err != nil {
		return nil, err
	}
	defer listenerFile.Close()
	newListener := Listener{
		Address:  address,
		FD:       3,
		Filename: listenerFile.Name(),
	}
	listenerEnv, err := json.Marshal(newListener)
	if err != nil {
		return nil, err
	}

	// Pass stdin, stdout, and stderr along with the _listener to the child.
	// Os only allows to pass these
	files := []*os.File{
		os.Stdin,
		os.Stdout,
		os.Stderr,
		listenerFile,
	}

	// Get current environment and add in the _listener to it.
	environment := append(os.Environ(), "LISTENER="+string(listenerEnv))

	// Get current process name and directory.
	execName, err := os.Executable()
	if err != nil {
		return nil, err
	}
	execDir := filepath.Dir(execName)
	log.Println(execName)

	// Spawn child process.
	p, err := os.StartProcess(execName, []string{execName}, &os.ProcAttr{
		Dir:   execDir,
		Env:   environment,
		Files: files,
		Sys:   &syscall.SysProcAttr{},
	})
	if err != nil {
		return nil, err
	}

	return p, nil
}

func WaitForSignals(address string, listener net.Listener, server *http.Server) error {
	signalCh := make(chan os.Signal, 1024)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGQUIT)
	for {
		select {
		case s := <-signalCh:
			log.Printf("[%v] signal received.\n", s)
			switch s {
			case syscall.SIGHUP:
				// Fork a child process.
				p, err := ForkChild(address, listener)
				if err != nil {
					log.Printf("Unable to fork child process: %v.\n", err)
					continue
				}
				log.Printf("Forked child process %v.\n", p.Pid)

				// Create a context that will expire in 5 seconds and use this as a
				// timeout to Shutdown.
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// Return any errors during shutdown.
				return server.Shutdown(ctx)
			case syscall.SIGUSR2:
				// Fork a child process.
				p, err := ForkChild(address, listener)
				if err != nil {
					log.Printf("Unable to fork child process: %v.\n", err)
					continue
				}

				// Print the PID of the forked process and keep waiting for more signals.
				log.Printf("Forked child process %v.\n", p.Pid)
			case syscall.SIGINT, syscall.SIGQUIT:
				// Create a context that will expire in 5 seconds and use this as a
				// timeout to Shutdown.
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// Return any errors during shutdown.
				return server.Shutdown(ctx)
			}
		}
	}
}