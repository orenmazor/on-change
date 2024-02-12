/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/fatih/color"
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

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	// Start listening for events.
	// TODO this needs to be buffered so we're only running things once
	// otherwise if VIM does a bunch of actions to the file, or we save 5 files, we're going
	// to get an event for each activity, which sucks if the command to run is expensive
	go func() {
		for {
			select {
			case event, _ := <-watcher.Events:
				if !event.Has(fsnotify.Create) && !event.Has(fsnotify.Write) {
					continue
				}
				slog.Info("event received!", "event", event)
				rerunCommand(command)

			case err := <-watcher.Errors:
				slog.Error(err.Error())
				return
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

func rerunCommand(command string) {
	slog.Info("Rerunning", "command", command)

	command = "-c " + command
	args := strings.Split(command, " ")

	runnable := exec.Command("bash", args...)

	var stdout, stderr bytes.Buffer
	runnable.Stdout = &stdout
	runnable.Stderr = &stderr

	err := runnable.Run()

	if err != nil {
		fmt.Println("Error:", err)
	}

	color.Red("Stderr:\n\n")
	fmt.Println(stderr.String())
	color.Green("Stdout:\n\n")
	fmt.Println(stdout.String())
}
