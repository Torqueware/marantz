package serialport

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/joshuarubin/chanpubsub"

	serial "github.com/tarm/goserial"
)

type SerialPort struct {
	*chanpubsub.ChanPubSub
	Write  chan<- string
	Config serial.Config
	port   io.ReadWriteCloser
	opened bool
}

func (s *SerialPort) reader() {
	rd := bufio.NewReader(s.port)

	for {
		str, err := rd.ReadString('\r')

		if err != nil {
			log.Println("SerialPort::reader error", err)
		}

		str = strings.Trim(str, "@\r")
		log.Println("SERIAL ->", str)
		s.Pub(str)
	}
}

func (s *SerialPort) writer() chan<- string {
	ch := make(chan string)

	go func() {
		for str := range ch {
			str = strings.ToUpper(str)
			log.Println("SERIAL <-", str)
			str = fmt.Sprintf("@%s\r", str)
			_, err := s.port.Write([]byte(str))
			if err != nil {
				log.Println("SerialPort::writer error", err)
			}
		}
	}()

	return ch
}

func (s *SerialPort) Open() (err error) {
	if s.opened {
		return nil
	}

	s.port, err = serial.OpenPort(&s.Config)
	if err != nil {
		return err
	}

	if s.ChanPubSub == nil {
		s.ChanPubSub = chanpubsub.New()
	}

	go s.reader()

	if s.Write == nil {
		s.Write = s.writer()
	}

	s.opened = true
	return nil
}
