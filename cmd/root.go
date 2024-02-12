/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"log/slog"
	"os"
	"os/exec"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var path string
var command string

var rootCmd = &cobra.Command{
	Use:   "on-change",
	Short: "Run a command (-c) on changes to path (-p)",
	Run:   Monitor,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&path, "path", "p", ".", "Path to monitor for changes")
	rootCmd.Flags().StringVarP(&command, "command", "c", "uname", "Command to execute on changes")
}

func Monitor(cmd *cobra.Command, args []string) {
	slog.Info("Hello! Watching 👀", "path", path, "command", command)

	verifyPathExists(path)
	verifyCommandExists(command)

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, _ := <-watcher.Events:
				slog.Info("event received!", "event", event)
				if event.Has(fsnotify.Write) {
					log.Println("modified file:", event.Name)
				}
			}
		}
	}()

	// Add a path.
	err = watcher.Add(path)
	if err != nil {
		panic(err)
	}

	<-make(chan struct{})
}

func verifyPathExists(path string) {
	_, err := os.Stat(path)

	if err != nil {
		panic(err)
	}
}

func verifyCommandExists(command string) {
	_, err := exec.LookPath(command)

	if err != nil {
		panic(err)
	}
}
