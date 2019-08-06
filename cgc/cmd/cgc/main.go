package cgc

func main() {

	// create assets and wrap in process.
	config := readConfig()
	ch := make(chan string)
	go launchAssets(ch)

	// let system run for specified time, then shutdown
}

func launchAssets(ch chan string) {

}
