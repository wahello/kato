// Copyright (C) 2021 Gridworkz Co., Ltd.
// KATO, Application Management Platform

// Permission is hereby granted, free of charge, to any person obtaining a copy of this 
// software and associated documentation files (the "Software"), to deal in the Software
// without restriction, including without limitation the rights to use, copy, modify, merge,
// publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons 
// to whom the Software is furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all copies or 
// substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, 
// INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR
// PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE
// FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/gridworkz/kato/discover"
	"github.com/gridworkz/kato/eventlog/cluster"
	"github.com/gridworkz/kato/eventlog/conf"
	"github.com/gridworkz/kato/eventlog/db"
	"github.com/gridworkz/kato/eventlog/entry"
	"github.com/gridworkz/kato/eventlog/exit/web"
	"github.com/gridworkz/kato/eventlog/store"
	etcdutil "github.com/gridworkz/kato/util/etcd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// LogServer -
type LogServer struct {
	Conf         conf.Conf
	Entry        *entry.Entry
	Logger       *logrus.Logger
	SocketServer *web.SocketServer
	Cluster      cluster.Cluster
}

// NewLogServer creates a new NewLogServer.
func NewLogServer() *LogServer {
	conf := conf.Conf{}
	return &LogServer{
		Conf: conf,
	}
}

//AddFlags
func (s *LogServer) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Conf.Entry.EventLogServer.BindIP, "eventlog.bind.ip", "0.0.0.0", "Collect the log service to listen the IP")
	fs.IntVar(&s.Conf.Entry.EventLogServer.BindPort, "eventlog.bind.port", 6366, "Collect the log service to listen the Port")
	fs.IntVar(&s.Conf.Entry.EventLogServer.CacheMessageSize, "eventlog.cache", 100, "the event log server cache the receive message size")
	fs.StringVar(&s.Conf.Entry.DockerLogServer.BindIP, "dockerlog.bind.ip", "0.0.0.0", "Collect the log service to listen the IP")
	fs.StringVar(&s.Conf.Entry.DockerLogServer.Mode, "dockerlog.mode", "stream", "the server mode zmq or stream")
	fs.IntVar(&s.Conf.Entry.DockerLogServer.BindPort, "dockerlog.bind.port", 6362, "Collect the log service to listen the Port")
	fs.IntVar(&s.Conf.Entry.DockerLogServer.CacheMessageSize, "dockerlog.cache", 200, "the docker log server cache the receive message size")
	fs.StringSliceVar(&s.Conf.Entry.MonitorMessageServer.SubAddress, "monitor.subaddress", []string{"tcp://127.0.0.1:9442"}, "monitor message source address")
	fs.IntVar(&s.Conf.Entry.MonitorMessageServer.CacheMessageSize, "monitor.cache", 200, "the monitor sub server cache the receive message size")
	fs.StringVar(&s.Conf.Entry.MonitorMessageServer.SubSubscribe, "monitor.subscribe", "ceptop", "the monitor message sub server subscribe info")
	fs.BoolVar(&s.Conf.ClusterMode, "cluster", true, "Whether open cluster mode")
	fs.StringVar(&s.Conf.Cluster.Discover.InstanceIP, "cluster.instance.ip", "", "The current instance IP in the cluster can be communications.")
	fs.StringVar(&s.Conf.Cluster.Discover.Type, "discover.type", "etcd", "the instance in cluster auto discover way.")
	fs.StringSliceVar(&s.Conf.Cluster.Discover.EtcdAddr, "discover.etcd.addr", []string{"http://127.0.0.1:2379"}, "set all etcd server addr in cluster for message instence auto discover.")
	fs.StringVar(&s.Conf.Cluster.Discover.EtcdCaFile, "discover.etcd.ca", "", "verify etcd certificates of TLS-enabled secure servers using this CA bundle")
	fs.StringVar(&s.Conf.Cluster.Discover.EtcdCertFile, "discover.etcd.cert", "", "identify secure etcd client using this TLS certificate file")
	fs.StringVar(&s.Conf.Cluster.Discover.EtcdKeyFile, "discover.etcd.key", "", "identify secure etcd client using this TLS key file")
	fs.StringVar(&s.Conf.Cluster.Discover.HomePath, "discover.etcd.homepath", "/event", "etcd home key")
	fs.StringVar(&s.Conf.Cluster.Discover.EtcdUser, "discover.etcd.user", "", "etcd server user info")
	fs.StringVar(&s.Conf.Cluster.Discover.EtcdPass, "discover.etcd.pass", "", "etcd server user password")
	fs.StringVar(&s.Conf.Cluster.PubSub.PubBindIP, "cluster.bind.ip", "0.0.0.0", "Cluster communication to listen the IP")
	fs.IntVar(&s.Conf.Cluster.PubSub.PubBindPort, "cluster.bind.port", 6365, "Cluster communication to listen the Port")
	fs.StringVar(&s.Conf.EventStore.MessageType, "message.type", "json", "Receive and transmit the log message type.")
	fs.StringVar(&s.Conf.EventStore.GarbageMessageSaveType, "message.garbage.save", "file", "garbage message way of storage")
	fs.StringVar(&s.Conf.EventStore.GarbageMessageFile, "message.garbage.file", "/var/log/envent_garbage_message.log", "save garbage message file path when save type is file")
	fs.Int64Var(&s.Conf.EventStore.PeerEventMaxLogNumber, "message.max.number", 100000, "the max number log message for peer event")
	fs.IntVar(&s.Conf.EventStore.PeerEventMaxCacheLogNumber, "message.cache.number", 256, "Maintain log the largest number in the memory peer event")
	fs.Int64Var(&s.Conf.EventStore.PeerDockerMaxCacheLogNumber, "dockermessage.cache.number", 128, "Maintain log the largest number in the memory peer docker service")
	fs.IntVar(&s.Conf.EventStore.HandleMessageCoreNumber, "message.handle.core.number", 2, "The number of concurrent processing receive log data.")
	fs.IntVar(&s.Conf.EventStore.HandleSubMessageCoreNumber, "message.sub.handle.core.number", 3, "The number of concurrent processing receive log data. more than message.handle.core.number")
	fs.IntVar(&s.Conf.EventStore.HandleDockerLogCoreNumber, "message.dockerlog.handle.core.number", 2, "The number of concurrent processing receive log data. more than message.handle.core.number")
	fs.StringVar(&s.Conf.Log.LogLevel, "log.level", "info", "app log level")
	fs.StringVar(&s.Conf.Log.LogOutType, "log.type", "stdout", "app log output type. stdout or file ")
	fs.StringVar(&s.Conf.Log.LogPath, "log.path", "/var/log/", "app log output file path.it is effective when log.type=file")
	fs.StringVar(&s.Conf.WebSocket.BindIP, "websocket.bind.ip", "0.0.0.0", "the bind ip of websocket for push event message")
	fs.IntVar(&s.Conf.WebSocket.BindPort, "websocket.bind.port", 6363, "the bind port of websocket for push event message")
	fs.IntVar(&s.Conf.WebSocket.SSLBindPort, "websocket.ssl.bind.port", 6364, "the ssl bind port of websocket for push event message")
	fs.BoolVar(&s.Conf.WebSocket.EnableCompression, "websocket.compression", true, "weither enable compression for web socket")
	fs.IntVar(&s.Conf.WebSocket.ReadBufferSize, "websocket.readbuffersize", 4096, "the readbuffersize of websocket for push event message")
	fs.IntVar(&s.Conf.WebSocket.WriteBufferSize, "websocket.writebuffersize", 4096, "the writebuffersize of websocket for push event message")
	fs.IntVar(&s.Conf.WebSocket.MaxRestartCount, "websocket.maxrestart", 5, "the max restart count of websocket for push event message")
	fs.BoolVar(&s.Conf.WebSocket.SSL, "websocket.ssl", false, "whether to enable websocket  SSL")
	fs.StringVar(&s.Conf.WebSocket.CertFile, "websocket.certfile", "/etc/ssl/gridworkz.com/gridworkz.com.crt", "websocket ssl cert file")
	fs.StringVar(&s.Conf.WebSocket.KeyFile, "websocket.keyfile", "/etc/ssl/gridworkz.com/gridworkz.com.key", "websocket ssl cert file")
	fs.StringVar(&s.Conf.WebSocket.TimeOut, "websocket.timeout", "1m", "Keep websocket service the longest time when without message ")
	fs.StringVar(&s.Conf.WebSocket.PrometheusMetricPath, "monitor-path", "/metrics", "promethesu monitor metrics path")
	fs.StringVar(&s.Conf.EventStore.DB.Type, "db.type", "mysql", "Data persistence type.")
	fs.StringVar(&s.Conf.EventStore.DB.URL, "db.url", "root:admin@tcp(127.0.0.1:3306)/event", "Data persistence db url.")
	fs.IntVar(&s.Conf.EventStore.DB.PoolSize, "db.pool.size", 3, "Data persistence db pool init size.")
	fs.IntVar(&s.Conf.EventStore.DB.PoolMaxSize, "db.pool.maxsize", 10, "Data persistence db pool max size.")
	fs.StringVar(&s.Conf.EventStore.DB.HomePath, "docker.log.homepath", "/grdata/logs/", "container log persistent home path")
	fs.StringVar(&s.Conf.Entry.NewMonitorMessageServerConf.ListenerHost, "monitor.udp.host", "0.0.0.0", "receive new monitor udp server host")
	fs.IntVar(&s.Conf.Entry.NewMonitorMessageServerConf.ListenerPort, "monitor.udp.port", 6166, "receive new monitor udp server port")
	fs.StringVar(&s.Conf.Cluster.Discover.NodeID, "node-id", "", "the unique ID for this node.")
	fs.DurationVar(&s.Conf.Cluster.PubSub.PollingTimeout, "zmq4-polling-timeout", 200*time.Millisecond, "The timeout determines the time-out on the polling of sockets")
}

//InitLog
func (s *LogServer) InitLog() {
	log := logrus.New()
	if l, err := logrus.ParseLevel(s.Conf.Log.LogLevel); err == nil {
		log.Level = l
	} else {
		logrus.Warning("log.level is not valid.will set it is 'info'")
	}
	switch s.Conf.Log.LogOutType {
	case "stdout":
		log.Out = os.Stdout
	case "file":
		file, err := os.Stat(s.Conf.Log.LogPath)
		if err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(s.Conf.Log.LogPath, os.ModeDir); err != nil {
					logrus.Errorf("Create log dir error.%s,The log configuration does not take effect", err.Error())
				}
			}
		}
		if file.IsDir() {
			file, err := os.OpenFile(path.Join(s.Conf.Log.LogPath, "event_log.log"), os.O_RDWR|os.O_WRONLY, 755)
			if err != nil {
				logrus.Errorf("Open log file error.%s,The log configuration does not take effect", err.Error())
			} else {
				log.Out = file
			}
		} else { //Use the specified file path directly
			file, err := os.OpenFile(s.Conf.Log.LogPath, os.O_RDWR|os.O_WRONLY, 755)
			if err != nil {
				logrus.Errorf("Open log file error.%s,The log configuration does not take effect", err.Error())
			} else {
				log.Out = file
			}
		}
	}
	log.Formatter = &logrus.TextFormatter{}
	s.Logger = log
}

//InitConf
func (s *LogServer) InitConf() {
	s.Conf.Cluster.Discover.ClusterMode = s.Conf.ClusterMode
	s.Conf.Cluster.PubSub.ClusterMode = s.Conf.ClusterMode
	s.Conf.EventStore.ClusterMode = s.Conf.ClusterMode
	s.Conf.Cluster.Discover.DockerLogPort = s.Conf.Entry.DockerLogServer.BindPort
	s.Conf.Cluster.Discover.WebPort = s.Conf.WebSocket.BindPort
	if os.Getenv("MYSQL_HOST") != "" && os.Getenv("MYSQL_USER") != "" && os.Getenv("MYSQL_PASSWORD") != "" {
		s.Conf.EventStore.DB.URL = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"),
			os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_PORT"), os.Getenv("MYSQL_DATABASE"))
	}
	if os.Getenv("CLUSTER_BIND_IP") != "" {
		s.Conf.Cluster.PubSub.PubBindIP = os.Getenv("CLUSTER_BIND_IP")
	}
}

//Run
func (s *LogServer) Run() error {
	s.Logger.Debug("Start run server.")
	log := s.Logger

	etcdClientArgs := &etcdutil.ClientArgs{
		Endpoints: s.Conf.Cluster.Discover.EtcdAddr,
		CaFile:    s.Conf.Cluster.Discover.EtcdCaFile,
		CertFile:  s.Conf.Cluster.Discover.EtcdCertFile,
		KeyFile:   s.Conf.Cluster.Discover.EtcdKeyFile,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	etcdClient, err := etcdutil.NewClient(ctx, etcdClientArgs)
	if err != nil {
		return err
	}

	//init new db
	if err := db.CreateDBManager(s.Conf.EventStore.DB); err != nil {
		logrus.Infof("create db manager error, %v", err)
		return err
	}

	storeManager, err := store.NewManager(s.Conf.EventStore, log.WithField("module", "MessageStore"))
	if err != nil {
		return err
	}
	healthInfo := storeManager.HealthCheck()
	if err := storeManager.Run(); err != nil {
		return err
	}
	defer storeManager.Stop()
	if s.Conf.ClusterMode {
		s.Cluster = cluster.NewCluster(etcdClient, s.Conf.Cluster, log.WithField("module", "Cluster"), storeManager)
		if err := s.Cluster.Start(); err != nil {
			return err
		}
		defer s.Cluster.Stop()
	}
	s.SocketServer = web.NewSocket(s.Conf.WebSocket, s.Conf.Cluster.Discover, etcdClient,
		log.WithField("module", "SocketServer"), storeManager, s.Cluster, healthInfo)
	if err := s.SocketServer.Run(); err != nil {
		return err
	}
	defer s.SocketServer.Stop()

	s.Entry = entry.NewEntry(s.Conf.Entry, log.WithField("module", "EntryServer"), storeManager)
	if err := s.Entry.Start(); err != nil {
		return err
	}
	defer s.Entry.Stop()

	//Service registration
	grpckeepalive, err := discover.CreateKeepAlive(etcdClientArgs, "event_log_event_grpc",
		s.Conf.Cluster.Discover.NodeID, s.Conf.Cluster.Discover.InstanceIP, s.Conf.Entry.EventLogServer.BindPort)
	if err != nil {
		return err
	}
	if err := grpckeepalive.Start(); err != nil {
		return err
	}
	defer grpckeepalive.Stop()

	udpkeepalive, err := discover.CreateKeepAlive(etcdClientArgs, "event_log_event_udp",
		s.Conf.Cluster.Discover.NodeID, s.Conf.Cluster.Discover.InstanceIP, s.Conf.Entry.NewMonitorMessageServerConf.ListenerPort)
	if err != nil {
		return err
	}
	if err := udpkeepalive.Start(); err != nil {
		return err
	}
	defer udpkeepalive.Stop()

	httpkeepalive, err := discover.CreateKeepAlive(etcdClientArgs, "event_log_event_http",
		s.Conf.Cluster.Discover.NodeID, s.Conf.Cluster.Discover.InstanceIP, s.Conf.WebSocket.BindPort)
	if err != nil {
		return err
	}
	if err := httpkeepalive.Start(); err != nil {
		return err
	}
	defer httpkeepalive.Stop()

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case <-term:
		log.Warn("Received SIGTERM, exiting gracefully...")
	case err := <-s.SocketServer.ListenError():
		log.Errorln("Error listen web socket server, exiting gracefully:", err)
	case err := <-storeManager.Error():
		log.Errorln("Store receive a error, exiting gracefully:", err)
	}
	log.Info("See you next time!")
	return nil
}
