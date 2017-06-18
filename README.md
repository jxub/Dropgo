# Dropgo

JSON API for exposing the filesystem written in Golang, complete with the authentication and loggin middelware and login, logout and internal templates. API to be consumed client-side with AJAX, with api routes defined in main.go

### Run

To run in Linux, make a `git clone https://github.com/jxub/Dropgo`
then, `cd Dropgo/bin` into the compiled program and add executable bit to the binary: `chmod +x dropgo`
Run the app by simply typing `./dropgo`, and that's about it!

### Build

To build the most recent vesion, go to src/ with `cd src/` and type `go build -o ../bin/dropgo`, then follow the previous steps `cd ../bin`, `chmod +x dropgo` to add permissions, finally `./dropgo` to run. You must have installed Golang 1.8 or higher. 

### TODO

Consume the api with jQuery/VueJS/React on the client. But ughhhh, JS :/
