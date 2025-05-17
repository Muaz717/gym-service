package grpcerrors

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

func ParseValidationError(err error) string {
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		return err.Error()
	}

	for _, detail := range st.Details() {
		switch t := detail.(type) {
		case *errdetails.BadRequest:
			var msgs []string
			for _, v := range t.FieldViolations {
				msgs = append(msgs, v.Field+": "+v.Description)
			}
			return strings.Join(msgs, "; ")
		}
	}

	// fallback
	return st.Message()
}
