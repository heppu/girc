package girc

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
)

const (
	PASS = "PASS *"
	NICK = "NICK %s"
	USER = "USER %s 0 * :%s"
)

type Client struct {
	conn     *net.TCPConn
	server   string
	nick     string
	realName string
	rx       chan []byte
	tx       chan string
	quit     chan bool
}

func NewClient(s string, n string, r string) (*Client, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", s)
	if err != nil {
		return nil, err
	}

	c, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	rx := make(chan []byte)
	tx := make(chan string)
	q := make(chan bool)

	newClient := Client{c, s, n, r, rx, tx, q}

	go newClient.handleMessageFromServer()
	go newClient.listenServer()

	return &newClient, nil
}

func (c *Client) listenServer() {
	for {
		message, isPrefix, err := bufio.NewReader(c.conn).ReadLine()
		fmt.Printf("isPefix: %v\n", isPrefix)
		if err != nil {
			fmt.Println(err)
			c.Quit()
			break
		}
		c.rx <- message
	}
}

func (c *Client) handleMessageFromServer() {
L:
	for {
		select {
		case msg := <-c.rx:
			fmt.Printf("SERVER: %s\n", msg)
			c.parseMsg(msg)
		case <-c.quit:
			fmt.Println("Closing client")
			break L
		}
	}
	c.quit <- true
}

func (c *Client) Connect() {
	fmt.Println("CONNECT")
	c.write(PASS)
	c.write(fmt.Sprintf(NICK, c.nick))
	c.write(fmt.Sprintf(USER, c.nick, c.realName))
}

func (c *Client) Join() {
	fmt.Println("JOIN")
	c.write("JOIN #gobot")
}

func (c *Client) Send() {
	fmt.Println("PRIVMSG")
	c.write("PRIVMSG #gobot hello")
}

func (c *Client) Quit() {
	fmt.Println("QUIT")
	c.quit <- true
	<-c.quit
	fmt.Println("Quit done")
	return
}

func (c *Client) write(msg string) {
	fmt.Fprintf(c.conn, "%s\r\n", msg)
}

func (c *Client) parseMsg(msg []byte) error {
	if len(msg) == 0 {
		return errors.New("EOF")
	}
	str := string(msg[:len(msg)])

	parameters := strings.Split(str, ":")
	fmt.Printf("Parsed: %+v\n", parameters)
	return nil
}
