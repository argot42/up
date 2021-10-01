package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"io"
	"path/filepath"
)

var html string = `<!DOCTYPE html>
<html>
	<head>
		<title>☯</title>
		<style>
			.center {
				display: flex;
				flex-direction: column;
				align-items: center;
				justify-content: center;
				text-align: center;
			}
			.send {
				margin-top: 1em;
			}
			.filepicker {
				text-decoration: none;
			}
		</style>
	</head>
	<body>
		<div class=center>
			<h1>アップロード</h1>
			<form method="post" enctype="multipart/form-data">
				<input class="filepicker" type="file" id="file" name="myFile" multiple><br>
				<button class="send">send</button>
			</form>
		</div>
	</body>
</html>
`

func main() {
	if len(os.Args)	< 3 {
		fmt.Printf("usage: %s <host>:<port> <dir>\n", os.Args[0])
		return
	}

	logErr := log.New(os.Stderr, "", log.LstdFlags)

	socket := os.Args[1]
	dir := os.Args[2]

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			fmt.Fprintf(w, html)
		case "POST":
			reader, err := r.MultipartReader()
			if err != nil {
				logErr.Fatal(err)
			}

			var errors []string
			var file *os.File

			for {
				part, err := reader.NextPart()
				if err != nil {
					if err != io.EOF {
						logErr.Fatal(err)
					}
					break
				}

				path := filepath.Join(dir, part.FileName())
				file, err = os.OpenFile(path, os.O_WRONLY | os.O_CREATE | os.O_EXCL, 0640)
				if err != nil {
					logErr.Print("open file: ", err)
					errors = append(errors, path)
					goto Clean
				}

				_, err = io.Copy(file, part)
				if err != nil {
					logErr.Print("copy: ", err)
					errors = append(errors, path)
				}

Clean:
				err = part.Close()
				if err != nil {
					logErr.Print("part close: ", err)
				}

				if file == nil {
					continue
				}

				err = file.Close()
				if err != nil {
					logErr.Print("file close: ", err)
				}
			}

			if len(errors) == 0 {
				http.Error(w, "All goin' good", http.StatusOK)
			} else {
				fmt.Fprintln(w, errors)
			}
		}
	})

	log.Fatal(http.ListenAndServe(socket, nil))
}
