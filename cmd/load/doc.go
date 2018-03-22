/*
Package main allows a caller to generate load on fizbus

Build the go binary
  go build -o loader

Setup a binder
  ./loader -mode=binder -binder.concurrency=5

Setup a sender
  ./loader -mode=sender -sender.count=20 -sender.think-time=10
This will run the binary in sender mode.

Options
  endpoint: rabbitmw URL, the default is amqp://guest:guest@192.168.99.100:5672/
  mode: either sender or binder
  sender.count: the number of sender(s). Each sender is a go routine which sends a request and waits for a reply.
  sender.think-time: think time in millisecond(s) for a sender. The amount of time it will wait before sending another request.
  binder.concurrency: The number of messages the server will try to keep on the network for each consumer.
*/
package main
