# go-webcrawler
[![Build Status](https://travis-ci.org/nladuo/go-webcrawler.svg)](https://travis-ci.org/nladuo/go-webcrawler)

a simple, concurrent , distributed, go-implemented web crawler framework.<br>

## Installation
```
go get github.com/nladuo/go-webcrawler  
```
## Dependencies
```
go get github.com/samuel/go-zookeeper
go get github.com/jinzhu/gorm
go get github.com/nladuo/DLocker
go get github.com/PuerkitoBio/goquery
```
## About the Modes of go-webcrawler
go-webcrawler is a simple web crawler framework to let you build concurrent and distributed web crawler application. There are three modes of go-webcrawler: local memory mode, local sql mode and distributed sql mode.
#### Local Memory Mode
In this mode, the framework would store the intermediate data directly into memory. If the url list's size of your web crawler application would not grow exponentially, or you PC's memory is big enough to utilize, you can use this mode.

#### Local Sql Mode
In this mode, the framework would store the intermediate data into a sql database. Because of using [an ORM framework](https://github.com/jinzhu/gorm) for database manipulation, you can use sqlite3, postgreSQL, mysql and so on... You would not worry about the the massive request urls running out of your PC's memory.

#### Distributed Sql Mode
Same as the Local Sql Mode, The Distributed Sql Mode would store the intermediate data into a sql database too. The difference between them is that the distributed one need zookeeper for coordination.You can check out the zookeeper configuration <a href="http://zookeeper.apache.org/doc/r3.4.6/zookeeperStarted.html">here</a>.

## Usage
see the example.