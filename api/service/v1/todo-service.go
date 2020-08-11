package v1

import (
	"context"
	"database/sql"
	"fmt"
	v1 "go-grpc/api/proto/v1"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	apiVersion = "v1"
)

type toDoServiceServer struct {
	db *sql.DB
}

func NewToDoServiceServer(db *sql.DB) *toDoServiceServer {
	return &toDoServiceServer{
		db: db,
	}
}

func (serviceServer *toDoServiceServer) checkAPI(api string) error {
	if len(api) > 0 {
		if apiVersion != api {
			return status.Error(codes.Unimplemented, fmt.Sprintf("unsupported API version:service implements api version is '%s', but given is '%s'", apiVersion, api))
		}
	}
	return nil
}

func (serviceServer *toDoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	conn, err := serviceServer.db.Conn(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "连接数据库失败."+err.Error())
	}
	return conn, nil
}

func (serviceServer *toDoServiceServer) Create(ctx context.Context, req *v1.CreateRequest) (*v1.CreateResponse, error) {
	if err := serviceServer.checkAPI(req.Api); err != nil {
		return nil, err
	}
	conn, err := serviceServer.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, err = ptypes.Timestamp(req.ToDo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "参数错误"+err.Error())
	}
	res, err := conn.ExecContext(ctx, "INSERT INTO TODO(`Title`, `Description`, `Reminder`) VALUES (?, ?, ?)", req.ToDo.Title, req.ToDo.Description, req.ToDo.Reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, "添加TODO失败"+err.Error())
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Unknown, "获取最新ID失败"+err.Error())
	}
	return &v1.CreateResponse{
		Api: apiVersion,
		Id:  id,
	}, nil
}

func (serviceServer *toDoServiceServer) Read(ctx context.Context, req *v1.ReadRequest) (*v1.ReadResponse, error) {
	if err := serviceServer.checkAPI(req.Api); err != nil {
		return nil, err
	}
	conn, err := serviceServer.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rows, err := conn.QueryContext(ctx, "SELECT `ID`, `Title`, `Description`, `Reminder` FROM TODO WHERE `ID` = ?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "查询失败"+err.Error())
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "获取数据失败"+err.Error())
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ID = '%d' 查无此人", req.Id))
	}

	var toDo v1.ToDo
	if err = rows.Scan(&toDo.Id, &toDo.Title, &toDo.Description, &toDo.Reminder); err != nil {
		return nil, status.Error(codes.Unknown, "查找数据失败"+err.Error())
	}
	if rows.Next() {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("用户: '%d' 查找到多条数据", req.Id))
	}
	return &v1.ReadResponse{
		Api:  apiVersion,
		ToDo: &toDo,
	}, nil
}

func (serviceServer *toDoServiceServer) Update(ctx context.Context, req *v1.UpdateRequest) (*v1.UpdateResponse, error) {
	if err := serviceServer.checkAPI(req.Api); err != nil {
		return nil, err
	}
	conn, err := serviceServer.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, err = ptypes.Timestamp(req.ToDo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "参数错误：reminder无效"+err.Error())
	}
	res, err := conn.ExecContext(ctx, "UPDATE TODO SET `Title=?`, `Description=?`, `Reminder=?` WHERE `ID` = ?",
		req.ToDo.Title, req.ToDo.Description, req.ToDo.Reminder, req.ToDo.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "更新失败"+err.Error())
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "受影响行数获取失败"+err.Error())
	}
	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("查无此人：'%d'", req.ToDo.Id))
	}
	return &v1.UpdateResponse{
		Api:     apiVersion,
		Updated: rows,
	}, nil
}

func (serviceServer *toDoServiceServer) Delete(ctx context.Context, req *v1.DeleteRequest) (*v1.DeleteResponse, error) {
	if err := serviceServer.checkAPI(req.Api); err != nil {
		return nil, err
	}
	conn, err := serviceServer.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	res, err := conn.ExecContext(ctx, "DELETE FROM TODO WHERE `ID = ?`", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "删除失败"+err.Error())
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "受影响行数获取失败"+err.Error())
	}
	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("查无此人：'%d'", req.Id))
	}
	return &v1.DeleteResponse{
		Api:     apiVersion,
		Deleted: rows,
	}, nil
}

func (serviceServer *toDoServiceServer) ReadAll(ctx context.Context, req *v1.ReadAllRequest) (*v1.ReadAllResponse, error) {
	if err := serviceServer.checkAPI(req.Api); err != nil {
		return nil, err
	}
	conn, err := serviceServer.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rows, err := conn.QueryContext(ctx, "SELECT `ID`, `Title`, `Description`, `Reminder` FROM TODO")
	if err != nil {
		return nil, status.Error(codes.Unknown, "查询失败"+err.Error())
	}
	defer rows.Close()
	var list []*v1.ToDo
	var reminder time.Time
	for rows.Next() {
		toDo := new(v1.ToDo)
		if err = rows.Scan(&toDo.Id, &toDo.Title, &toDo.Description, &reminder); err != nil {
			return nil, status.Error(codes.Unknown, "查询失败："+err.Error())
		}
		toDo.Reminder, err = ptypes.TimestampProto(reminder)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "reminder参数无效"+err.Error())
		}
		list = append(list, toDo)
	}
	if err = rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "获取数据失败"+err.Error())
	}
	return &v1.ReadAllResponse{
		Api:  apiVersion,
		ToDo: list,
	}, nil
}
