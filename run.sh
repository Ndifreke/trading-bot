#!/bin/bash
rungo () {
        if [ $# -eq 0 ]
                then nodemon --delay 2 --exec go run main.go --signal SIGTERM
        elif [ $# -eq 1 ]
                then nodemon --delay 2 --exec go run $1 --signal SIGTERM
        fi
}
rungo $1
