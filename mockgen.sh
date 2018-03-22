#!/bin/sh -x
PACKAGE="$GOPATH/src/github.com/tkashem/fizbus"


mockgen -destination "${PACKAGE}/mock_context_test.go" -package fizbus golang.org/x/net/context Context 
mockgen -destination "${PACKAGE}/mock_fizbus_test.go" -source="${PACKAGE}/fizbus.go" -package fizbus Sender,Handler,Receiver
mockgen -destination="${PACKAGE}/mock_upstream_test.go" -package fizbus github.com/tkashem/fizbus/upstream Dialer,Connection,Channel,Table
mockgen -destination="${PACKAGE}/mock_cancelfunc_test.go" -source="${PACKAGE}/export_test.go" -package fizbus canceller
mockgen -destination "${PACKAGE}/mock_acknowledger_test.go" -package fizbus "github.com/streadway/amqp" Acknowledger

# all internal interfaces in internal.go
mockgen -destination="${PACKAGE}/mock_internal_test.go" -source="${PACKAGE}/internal.go" -package fizbus \
    sender,busRunner,starter,initializer,errorChannel,handler,handlerFactory,requestMap,converter,queueConsumer,senderLoop
    
# todo: mockgen has a known issue, the mock_fizbus_test.go will NOT compile because mockgen generates an import path 
# with vendor folder, as shown below
#   amqp "github.com/tkashem/fizbus/vendor/github.com/streadway/amqp"
# Replace with 
#   amqp "github.com/streadway/amqp"
#
# We need to find a solution for it.   