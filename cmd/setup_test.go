package cmd

// import (
// 	"testing"
// 	"time"
// 	"log"
// 	"os"
// 	expect "github.com/Netflix/go-expect"
// )

// // func TestGetProjectPath(t *testing.T) {

// 	c, err := expect.NewConsole(expect.WithStdout(os.Stdout))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer c.Close()


// 	go func() {
// 		c.ExpectEOF()
// 	}()

// 	err = cmd.Start()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	time.Sleep(time.Second)
// 	c.Send("iHello world\x1b")
// 	time.Sleep(time.Second)
// 	c.Send("dd")
// 	time.Sleep(time.Second)
// 	c.SendLine(":q!")

// 	err = cmd.Wait()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }
