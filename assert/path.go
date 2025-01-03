package assert

// import (
	// "fmt"
	// "os"
	// "path/filepath"
// )

// // PathExists verifica que el path exista
// func PathExists(path string) {
	// absPath, err := filepath.Abs(path)
	// assert(err != nil, fmt.Sprintf("Invalid path format: %s", path))
	// _, err = os.Stat(absPath)
	// assert(err != nil, fmt.Sprintf("Path does not exists: %s", absPath))
// }

// // PathIsDir verifica que el path exista y sea un directorio
// func PathIsDir(path string) {
	// absPath, err := filepath.Abs(path)
	// ErrNil(err, ")


	// info, err := os.Stat(absPath)
	// if err != nil {
		// condition := info.IsDir()
		// errMsg := fmt.Sprintf("Path is a directory, expected a file: %s", absPath)
		// assert(condition, errMsg)
	// }

	// condition := info.IsDir()
	// errMsg := fmt.Sprintf("Path is a directory, expected a file: %s", absPath)
	// assert(condition, errMsg)
// }

// // PathIsFile verifica que el path exista y sea un archivo
// func PathIsFile(path string) {
	// absPath, err := filepath.Abs(path)
	// assert(err != nil, fmt.Sprintf("Invalid path format: %s", path))

	// info, err := os.Stat(absPath)
	// assert(err != nil, fmt.Sprintf("Path does not exists: %s", absPath))

	// condition := !info.IsDir()
	// errMsg := fmt.Sprintf("Path is a directory, expected a file: %s", absPath)
	// assert(condition, errMsg)
// }
