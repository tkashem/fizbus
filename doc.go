/*
Package fizbus allows a caller to interact with an AMQP broker and exercise request-reply pattern.

Purpose

All API calls to services like task, app as specified in Fiz must go through the fizbus package. A bus provides two modes of operation:
    Sender: A sender can send request(s) to a queue and can wait for the reply
    Binder: A binder binds to a queue and continue to server requests off the queue

If we look at the services we have in Fiz, we can categorize them as below:
    Task: A task is exclusively a binder
    App: An app is both a sender and a binder
    Api: An api gateway is exclusively a sender

fizbus abstracts all internal details of AMQP and provides the user with a simple interface to accomplish request-reply pattern.

    The bus also integrates with google's net/context.
    To know more about net/context please visit https://blog.golang.org/context

*/
package fizbus
