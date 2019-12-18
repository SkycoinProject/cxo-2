# CXO Node and Tracker Integration Testing

Test will emulate comunication between 5 CXO Nodes communicating with tracker.

Nodes will take following actions:

* Node 1 will subscribe to Node 2 and 4 and will publish one file
* Node 2 will subscribe to Node 3 and will push directory twice - first one file and add * second file on update. It will subscribe later so that Node 3 has time to do 1 publish before Node 2 subscribes
* Node 3 will subscibe to no other nodes and will publish first file and then directory on two other times (creating one file in first directory push and replacing it in the second push with two new files)
* Node 4 will subscribe to Node 2 and Node 3 and will not push at all
* Node 5 will subscribe to Node 1 and will push once empty directory

At the end this should leave us with the following received content:

* Node 1 should receive directory with two files from Node 2 and nothing from Node 4
* Node 2 should receive directory from Node 3 that contains only two files from the latest push (previous records should be deleted)
* Node 3 will receive no files since he doesn't subscribe
* Node 4 should receive same content from Node 3 as Node 2 did and receive directory with two files from Node 2
* Node 5 should receive one file from Node 1
