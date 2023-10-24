package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var mu sync.Mutex

func createNamedPipe(path string) error {
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return syscall.Mkfifo(path, 0666)
}

func FD_SET(p *syscall.FdSet, i int) {
	p.Bits[i/64] |= 1 << (uint(i) % 64)
}

func FD_ISSET(p *syscall.FdSet, i int) bool {
	return (p.Bits[i/64] & (1 << (uint(i) % 64))) != 0
}

func serveFromPipe(pipePath string, timeoutDuration time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		remoteAddr := r.RemoteAddr
		userAgent := r.Header.Get("User-Agent")
		log.Infof("Received request from %s with User-Agent %s", remoteAddr, userAgent)

		pipe, err := os.OpenFile(pipePath, os.O_RDONLY|syscall.O_NONBLOCK, 0666)
		if err != nil {
			log.Error("Error opening pipe", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer pipe.Close()

		fd := int(pipe.Fd())
		err = syscall.SetNonblock(fd, true)
		if err != nil {
			log.Error("Failed to set non-blocking mode", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		rfds := &syscall.FdSet{}
		FD_SET(rfds, fd)

		tv := syscall.Timeval{
			Sec:  int64(timeoutDuration.Seconds()),
			Usec: int64(timeoutDuration.Microseconds()) % 1e6,
		}

		_, err = syscall.Select(fd+1, rfds, nil, nil, &tv)
		if err != nil {
			log.Error("Select error", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !FD_ISSET(rfds, fd) {
			log.Error("Timeout reading from pipe")
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		data, err := ioutil.ReadAll(pipe)
		if err != nil {
			log.Error("Error reading from pipe", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var jsonObj interface{}
		if json.Unmarshal(data, &jsonObj) != nil {
			log.Error("Invalid JSON received from pipe")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func cleanupPipe(pipeName string) {
	log.Info("Cleaning up", "pipe", pipeName)
	if err := os.RemoveAll(pipeName); err != nil && !os.IsNotExist(err) {
		log.Error("Could not clear the pipe", "err", err)
	}
}

func main() {
	var host string
	var port int
	var timeout int

	var rootCmd = &cobra.Command{
		Use:   "pipey [PIPE_NAME]",
		Short: "HTTP server that reads from a named pipe and exposes it over HTTP",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pipeName := args[0]
			// Setup signal catching
			sigs := make(chan os.Signal, 1)

			// Register for interupt (Control+C) and SIGTERM (from kill)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			// Create a function to handle cleanup
			go func() {
				sig := <-sigs
				log.Infof("Received signal: %s", sig)
				cleanupPipe(pipeName)
				os.Exit(0)
			}()

			if err := createNamedPipe(pipeName); err != nil {
				log.Fatal("Failed to create named pipe", "err", err)
			}

			http.HandleFunc("/", serveFromPipe(pipeName, time.Duration(timeout)*time.Second))
			address := fmt.Sprintf("%s:%d", host, port)
			log.Infof("Server started on %s", address)
			log.Fatal(http.ListenAndServe(address, nil))
		},
	}

	rootCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Host address to bind to")
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 1, "Timeout duration in seconds")

	if err := rootCmd.Execute(); err != nil {
		log.Error("CLI execution failed", "err", err)
	}
}
