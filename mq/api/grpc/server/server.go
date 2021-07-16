// Copyright (C) 2021 Gridworkz Co., Ltd.
// KATO, Application Management Platform

// This program is free software: you can redistribute it and/or modify
// it under the terms of the MIT License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-MIT usage of Kato,
// one or multiple Commercial Licenses authorized by Gridworkz Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// MIT License for more details.

// You should have received a copy of the MIT License
// along with this program. If not, see <http://opensource.org/licenses/MIT>.

package server

import (
	"fmt"

	"github.com/gridworkz/kato/util"

	"github.com/gridworkz/kato/mq/api/grpc/pb"
	"github.com/gridworkz/kato/mq/api/mq"

	"github.com/sirupsen/logrus"

	context "golang.org/x/net/context"

	proto "github.com/golang/protobuf/proto"
	grpc1 "google.golang.org/grpc"
)

type mqServer struct {
	actionMQ mq.ActionMQ
}

func (s *mqServer) Enqueue(ctx context.Context, in *pb.EnqueueRequest) (*pb.TaskReply, error) {
	if in.Topic == "" || !s.actionMQ.TopicIsExist(in.Topic) {
		return nil, fmt.Errorf("topic %s is not support", in.Topic)
	}
	if in.Message.TaskId == "" {
		in.Message.TaskId = util.NewUUID()
	}
	message, err := proto.Marshal(in.Message)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	err = s.actionMQ.Enqueue(ctx, in.Topic, string(message))
	if err != nil {
		return nil, err
	}
	logrus.Debugf("task (%v) enqueue.", in.Message.TaskType)
	return &pb.TaskReply{
		Status: "success",
	}, nil
}
func (s *mqServer) Topics(ctx context.Context, in *pb.TopicRequest) (*pb.TaskReply, error) {
	return &pb.TaskReply{
		Status: "success",
		Topics: s.actionMQ.GetAllTopics(),
	}, nil
}

func (s *mqServer) Dequeue(ctx context.Context, in *pb.DequeueRequest) (*pb.TaskMessage, error) {
	if in.Topic == "" || !s.actionMQ.TopicIsExist(in.Topic) {
		return nil, fmt.Errorf("topic %s is not support", in.Topic)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	message, err := s.actionMQ.Dequeue(ctx, in.Topic)
	if err != nil {
		return nil, err
	}
	var task pb.TaskMessage
	err = proto.Unmarshal([]byte(message), &task)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("task (%s) dnqueue by (%s).", task.GetTaskType(), in.ClientHost)
	return &task, nil
}

//RegisterServer
func RegisterServer(server *grpc1.Server, actionMQ mq.ActionMQ) {
	pb.RegisterTaskQueueServer(server, &mqServer{actionMQ})
}
