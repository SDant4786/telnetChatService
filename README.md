# Telnet Chat Service

## Running
go run .

## Connect to telnet
telnet 127.0.0.1 8181

## Postman collection for HTTP stuff in main folder
    telnetChatService.postman_collection.json
    
## Unit tests
go test -v ./...

## Features Implemented
#### Telnet Chat
    Client interacts with CLI
    Supports mulit client connections
    Supports channels
    Supports PMs
    Supports ignoring messages from a user
    Supports help menu 
#### Http
    /submitMessage
        Allows for messaging to:
            All connected users
            Directly to channels
            PMS
    /getLogs
        Returns the contents of the log file
    /stats
        Returns stats about connected user, messages sent, open channels
#### Config details stored in config file
#### Full unit test coverage

## Approach
    1.Break down what was needed.
        Packages:
            One of Telnet Stuff
            One for HTTP stuff
            One for Config stuff
    2.Found an online example of a telnet chat service as a resource to build off of
        https://github.com/dbnegative/go-telnet-chatserver/tree/master
    3.Broke apart the different requirements into empty functions
        Telnet server functions:
            Init
            Create a user
            Read input from CLI
            Print Help Menu to user
        Telnet User functions:
            Go routine for reading the users input
            Go routine for getting messages from other users
            Function to handle command inputs
                This was done so that each command would have its own function 
                and theroetically, be easier to write unit tests for
            Quit the service
            List all channels
            List all users
            Create a channel
            Join a channel
            Leave a channel
            Ignore a user
            Unignore a user
            Send a pm
            Send a message into a channel
            List all channels a user is in
        HTTP server functions:
            Init
            Submit Message
            Get logs
            Get stats
        Config:
            Load config
    4. Coding
        Config + unit tests
        Telnet stuff + unit tests
        HTTP stuff + unit tests


### Issues
     The main issue I had with this was trying to allow the http part 
     to interact with the telnet server using the same functions as the 
     telnet users. The way the unit tests were done for the telnet stuff lead
     me to believe I could create an "http" user by dialing the tcp endpoint in 
     code and using that connection in the same way as the telnet unit tests,
     that did not work though. Reading from the conn variable through an http 
     request caused the read to block and never release. And writing to the 
     conn more than once did the same thing. I eventually had to scrap that idea
     to meet the delieverables instead of trying to expand upon it like I would
     have liked to.
     
     I would have liked for the telnet unit tests to be strucutred differently 
     but GOs tests are async and a lot of the command tests needed to be 
     executed in order.
     
     No known bugs at this time.