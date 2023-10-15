package tool

import (
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	pb "github.com/zs-dima/auth-service/internal/gen/proto"
)

func DbIdToString(id *pgtype.UUID) string {
	src := id.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", src[0:4], src[4:6], src[6:8], src[8:10], src[10:16])
}

func DbIdToId(id *pgtype.UUID) *uuid.UUID {
	uuid := uuid.MustParse(DbIdToString(id))
	return &uuid
}

func DbIdToRpcId(id *pgtype.UUID) *pb.UUID {
	uuid := uuid.MustParse(DbIdToString(id))
	return &pb.UUID{Value: uuid.String()}
}

func RpcIdToDbId(id *pb.UUID) *pgtype.UUID {
	return StringToDbId(id.Value)
}

func RpcIdToId(id *pb.UUID) *uuid.UUID {
	uuid := uuid.MustParse(id.Value)
	return &uuid
}

func IdToRpcId(id *uuid.UUID) *pb.UUID {
	return &pb.UUID{Value: id.String()}
}

func IdToDbId(id *uuid.UUID) *pgtype.UUID {
	return StringToDbId(id.String())
}

func StringToDbId(id string) *pgtype.UUID {
	data, err := parseUUID(id)
	if err != nil {
		return &pgtype.UUID{
			Bytes: [16]byte{},
			Valid: false,
		}
	}
	return &pgtype.UUID{
		Bytes: data,
		Valid: true,
	}
}

// parseUUID converts a string UUID in standard form to a byte array.
// https://pkg.go.dev/github.com/emicklei/pgtalk/convert
func parseUUID(src string) (dst [16]byte, err error) {
	switch len(src) {
	case 36:
		src = src[0:8] + src[9:13] + src[14:18] + src[19:23] + src[24:]
	case 32:
		// dashes already stripped, assume valid
	default:
		// assume invalid.
		return dst, fmt.Errorf("cannot parse UUID %v", src)
	}

	buf, err := hex.DecodeString(src)
	if err != nil {
		return dst, err
	}

	copy(dst[:], buf)
	return dst, err
}
