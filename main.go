package main

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
)


func writeMapToFile(m map[string]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to create file: %v", err)
	}
	defer file.Close()
	for key, value := range m {
		_, err := file.WriteString(fmt.Sprintf("%s %s\n", key, value))
		if err != nil {
			return fmt.Errorf("unable to write to file: %v", err)
		}
	}
	return nil
}

func removerepo() error {
	fmt.Println("Removing GitGo repository...")
	err := os.RemoveAll(".gitgo")
	if err != nil {
		return fmt.Errorf("unable to remove directory: %v", err)
	}
	return nil
}

func gitgoinit() error {
	fmt.Println("Initializing GitGo repository...")

	if _, err := os.Stat(".gitgo"); err == nil {
		return fmt.Errorf("repository already exists")
	}

	// create a gitgo directory
	err := os.Mkdir(".gitgo", 0755)
	if err != nil {
		return fmt.Errorf("unable to create directory: %v", err)
	}

	// create the objects directory
	err = os.Mkdir(".gitgo/objects", 0755)
	if err != nil {
		return fmt.Errorf("unable to create objects directory: %v", err)
	}

	// create the commits directory
	err = os.Mkdir(".gitgo/commits", 0755)
	if err != nil {
		return fmt.Errorf("unable to create commits directory: %v", err)
	}

	// create the tracking list
	file, err := os.Create(".gitgo/tracking")
	if err != nil {
		return fmt.Errorf("unable to create tracking file: %v", err)
	}
	defer file.Close()
	return nil
}

func add(args []string) error {
	fmt.Println("Adding files...", args)
	for _, arg := range args {
		fmt.Println("Adding file:", arg)
		// Calculate SHA1 hash of file contents
		fileContent, err := os.ReadFile(arg)
		if err != nil {
			return fmt.Errorf("unable to read file %s: %v", arg, err)
		}
		hash := fmt.Sprintf("%x", sha1.Sum(fileContent))

		// write the entry to the tracking file
		file, err := os.OpenFile(".gitgo/tracking", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("unable to open tracking file: %v", err)
		}
		defer file.Close()

		// Write file name and hash to tracking file
		_, err = file.WriteString(fmt.Sprintf("%s %s\n", arg, hash))
		if err != nil {
			return fmt.Errorf("unable to write to tracking file: %v", err)
		}
	}
	return nil
}

func commit() error {
	fmt.Println("Committing changes...")
	// find the number of the commit
	commitNumber := 0
	commitDir := ".gitgo/commits"

	// read the commits directory
	files, err := os.ReadDir(commitDir)
	if err != nil {
		return fmt.Errorf("unable to read directory: %v", err)
	}
	for _, file := range files {
		if file.IsDir() {
			commitNumber++
		}
	}

	fmt.Printf("Commit number: %d\n", commitNumber)

	tracking_file, err := os.Open(".gitgo/tracking")
	if err != nil {
		return fmt.Errorf("unable to open tracking file: %v", err)
	}
	defer tracking_file.Close()
	nb_commited_files := 0
	// read the tracking file line by line
	tracked_files := make(map[string]string) // filename -> hash
	var filename, hash string
	for {
		_, err := fmt.Fscanf(tracking_file, "%s %s\n", &filename, &hash)
		if err != nil {
			break
		}
		tracked_files[filename] = hash
		fmt.Printf("Committing %s\n", filename)
		// read the file content
		addedFile, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("unable to read file %s: %v", filename, err)
		}
		calculatedHash := fmt.Sprintf("%x", sha1.Sum(addedFile))
		if calculatedHash == hash {
			continue
		}
		
		// create a new directory for the commit			
		nb_commited_files++
		objectDir := filepath.Join(".gitgo/commits", fmt.Sprintf("%d", commitNumber))
		if err := os.MkdirAll(objectDir, 0755); err != nil {
			return fmt.Errorf("unable to create directory: %v", err)
		}
		fmt.Printf("Created directory %s\n", objectDir)

		// write the file content to a new file in the directory
		objectFile := filepath.Join(objectDir, filename)
		err = os.WriteFile(objectFile, addedFile, 0644)
		if err != nil {
			return fmt.Errorf("unable to write file: %v", err)
		}
		tracked_files[filename] = calculatedHash
	}

	writeMapToFile(tracked_files, ".gitgo/tracking")

	if nb_commited_files == 0 {
		fmt.Println("No changes to commit")
	} else {
		fmt.Printf("Committed %d files\n", nb_commited_files)
	}
	return nil
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Usage: gitgo <command>")
		fmt.Println("Available commands: init, remove, commit")
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "init":
		err = gitgoinit()
	case "remove":
		err = removerepo()
	case "add":
		err = add(os.Args[2:])
	case "commit":
		err = commit()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Available commands: init, add, remove, commit")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
