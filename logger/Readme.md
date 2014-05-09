This package exports logging APIs that can be used to log to a file or a remote server
See the package documentation at [GoDoc] http://godoc.org/github.com/couchbaselabs/retriever/logger

Usage:

'''
func main() {

        var rl logger.LogWriter
        const DEFAULT_MODULE = "Retriever"   
        // create new instance of logger
        rl, err := logger.NewLogger(DEFAULT_MODULE, logger.LevelInfo)
        if err != nil {
                fmt.Sprintf("Cannot intialize logger %s", err.Error())                                 
        }
        // set logging to file
        rl.SetFile("")
        // enable keys
        rl.EnableKeys([]string{DEFAULT_MODULE, "Logger", "Stats"}) 
        // Log Info message
        rl.LogInfo("", DEFAULT_MODULE, "Retriever Server started")
        // Change the default log path. trace logs will still go to /tmp
        err := rl.SetDefaultPath("/dev/shm")

        ....
}
'''
