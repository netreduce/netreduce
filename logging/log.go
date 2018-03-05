package logging

import "log"

type Log struct{}

func (l *Log) Println(a ...interface{}) {
	log.Println(a...)
}

func (l *Log) prefixed(prefix string, a ...interface{}) {
	l.Println(append([]interface{}{prefix}, a...)...)
}

func (l *Log) Error(a ...interface{}) {
	l.prefixed("[ERROR]", a...)
}
