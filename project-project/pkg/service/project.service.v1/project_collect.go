package project_service_v1

import (
	"context"
	"go.uber.org/zap"
	"project-common/encrypts"
	"project-common/errs"
	"project-grpc/project"
	"project-project/internal/data"
	"project-project/pkg/model"
	"strconv"
	"time"
)

// UpdateCollectProject 更新项目收藏状态。
// 该方法根据msg中的CollectType字段来决定是收藏还是取消收藏项目。
// 参数:
//
//	ctx - 上下文，用于传递请求范围的上下文信息。
//	msg - 包含收藏操作信息的ProjectRpcMessage对象。
//
// 返回值:
//
//	*project.CollectProjectResponse - 收藏操作的响应对象。
//	error - 错误对象，如果操作成功，则返回nil。
func (ps *ProjectService) UpdateCollectProject(ctx context.Context, msg *project.ProjectRpcMessage) (*project.CollectProjectResponse, error) {
	// 解密项目代码。
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	// 将解密后的项目代码字符串转换为int64类型。
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	// 创建一个带有超时的上下文，用于控制后续操作不超过2秒。
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // 确保在函数退出时取消上下文。

	var err error
	// 根据收藏类型执行相应的操作。
	if "collect" == msg.CollectType {
		// 构建项目收藏对象。
		pc := &data.ProjectCollection{
			ProjectCode: projectCode,
			MemberCode:  msg.MemberId,
			CreateTime:  time.Now().UnixMilli(),
		}
		// 保存项目收藏信息。
		err = ps.projectRepo.SaveProjectCollect(c, pc)
	}
	if "cancel" == msg.CollectType {
		// 取消项目收藏。
		err = ps.projectRepo.DeleteProjectCollect(c, msg.MemberId, projectCode)
	}
	// 如果发生错误，记录日志并返回错误。
	if err != nil {
		zap.L().Error("project UpdateCollectProject SaveProjectCollect error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 操作成功，返回空响应对象和nil错误。
	return &project.CollectProjectResponse{}, nil
}
