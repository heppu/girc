package girc

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	//"strings"
)

const (
	PASS = "PASS *"
	NICK = "NICK %s"
	USER = "USER %s 0 * :%s"
	PING = "PING"
	PONG = "PONG"

	prefix     byte = 0x3A // Prefix or last argument
	prefixUser byte = 0x21 // Username
	prefixHost byte = 0x40 // Hostname
	space      byte = 0x20 // Separator
)

type NullString struct {
	Empty  bool
	String string
}

type Message struct {
	Prefix   NullString
	Command  NullString
	Params   []string
	Trailing NullString
}

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

	go newClient.messageDispatcher()
	go newClient.serverListener()

	return &newClient, nil
}

func (c *Client) serverListener() {
	for {
		message, _, err := bufio.NewReader(c.conn).ReadLine()
		if err != nil {
			fmt.Println(err)
			c.Quit()
			break
		}
		c.rx <- message
	}
}

func (c *Client) messageDispatcher() {
L:
	for {
		select {
		case rawMsg := <-c.rx:
			fmt.Printf("\nSERVER: %s\n", rawMsg)
			m, err := parseMsg(rawMsg)
			if err != nil {
				fmt.Printf("\n\n%v\n\n", err)
			} else {
				go c.handleMessage(m)
			}
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

func (c *Client) Privmsg() {
	fmt.Println("PRIVMSG")
	c.write("PRIVMSG #gobot hello")
}

func (c *Client) Pong() {
	fmt.Println("PONG")
	//c.write("PONG")
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

func (c *Client) handleMessage(m *Message) {
	if m.Command.String == PING {
		c.Pong()
	}
}

func parseMsg(rawMsg []byte) (parsedMessage *Message, err error) {
	m := Message{}
	parsedMessage = &m
	// Bypass empty messages
	if len(rawMsg) == 0 {
		err = errors.New("Empty message")
		m.Trailing = NullString{true, string(rawMsg)}
		return
	}

	i := 0
	rawLen := len(rawMsg)

	// Parse prefix
	if rawMsg[i] == prefix {
		i++
		for {
			// Malformed message, no command
			if i == rawLen {
				err = errors.New("No command")
				m.Trailing = NullString{true, string(rawMsg)}
				return
			}
			// Reached the end of prefix
			if rawMsg[i] == space {
				// TODO: How this should be handled
				// Empty prefix
				if i < 2 {
					err = errors.New("ERR: No command")
					m.Trailing = NullString{true, string(rawMsg)}
					return
				}
				break
			}
			i++
		}
		m.Prefix = NullString{true, string(rawMsg[1:i])}
		i++
	}
	fmt.Printf(" Prefix: '%s'\n", m.Prefix.String)

	// Parse command
	prevIndex := i
	for {
		// Malformed message
		if i == rawLen {
			err = errors.New("ERR: No space at the end of command")
			return
		}
		// Parsed command
		if rawMsg[i] == space {
			break
		}
		i++
	}
	m.Command = NullString{true, string(rawMsg[prevIndex:i])}
	fmt.Printf(" Command: '%s'\n", m.Command.String)
	i++

	// End of message
	if i == rawLen {
		fmt.Println("No args, no trailing")
		return
	}

	// Parse Params
	prevIndex = i

	for {
		// Malformed message
		if i == rawLen {
			err = errors.New("ERR: Malformed message")
			return
		}
		// Last parameter
		if rawMsg[i] == prefix {
			i++
			break
		}
		// Parsed parameter
		if rawMsg[i] == space {
			m.Params = append(m.Params, string(rawMsg[prevIndex:i]))
			prevIndex = i
		}
		i++
	}
	fmt.Printf(" Params(%d): %v\n", len(m.Params), m.Params)

	// Parse trailing
	if i < rawLen {
		m.Trailing = NullString{true, string(rawMsg[i:])}
	}
	fmt.Printf(" Trailing: '%v'\n", m.Trailing.String)

	return
}
