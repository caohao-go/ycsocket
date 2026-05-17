package proto

import (
	"server_golang/common/consts"
)

// QueryParams 查询参数
type QueryParams struct {
	Type      int8 // 查询类型 1-app 查询 2-web 查询
	Header    *RequestHeader
	WebHeader *WebReqHeader
}

func (q *QueryParams) Clone() *QueryParams {
	newParams := &QueryParams{
		Type: q.Type,
	}

	if q.Header != nil {
		newParams.Header = &RequestHeader{
			Version:     q.Header.Version,
			RequestType: q.Header.RequestType,
			RequestId:   q.Header.RequestId,
			TraceId:     q.Header.TraceId,
			Timestamp:   q.Header.Timestamp,
			Timeout:     q.Header.Timeout,
			Caller:      q.Header.Caller,
			Callee:      q.Header.Callee,
			Appid:       q.Header.Appid,
			Ip:          q.Header.Ip,
			Bak:         q.Header.Bak,
		}
	}

	if q.WebHeader != nil {
		newParams.WebHeader = &WebReqHeader{
			Version:   q.WebHeader.Version,
			RequestId: q.WebHeader.RequestId,
			Timestamp: q.WebHeader.Timestamp,
			Ip:        q.WebHeader.Ip,
			Tenant:    q.WebHeader.Tenant,
			UserIdentity: &UserIdentity{
				Userid:   q.WebHeader.UserIdentity.Userid,
				Account:  q.WebHeader.UserIdentity.Account,
				Nickname: q.WebHeader.UserIdentity.Nickname,
			},
		}
	}

	return newParams
}

func (q *QueryParams) GetUserIdentity() *UserIdentity {
	if q.WebHeader != nil {
		return q.WebHeader.UserIdentity
	}
	return nil
}

// GetQueryMode 根据 units 获取 query mode.
func GetQueryMode(units []*Unit) int8 {
	var unitNum int
	for _, unit := range units {
		if len(unit.Sub) > 0 {
			return consts.QueryModeCompound
		}

		if len(unit.Trans) > 0 {
			for _, transUnit := range unit.Trans {
				if len(transUnit.Sub) > 0 {
					return consts.QueryModeCompound
				}
			}

			unitNum = unitNum + len(unit.Trans)
		} else {
			unitNum++
		}
	}

	if unitNum <= 1 {
		return consts.QueryModeSingle
	} else {
		return consts.QueryModeParallel
	}
}
