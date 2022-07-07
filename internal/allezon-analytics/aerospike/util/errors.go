package errors

import ast "github.com/aerospike/aerospike-client-go/types"

func isError(err error, resultCode ast.ResultCode) bool {
	if err == nil {
		return false
	}

	aerr, ok := err.(ast.AerospikeError)
	return ok && aerr.ResultCode() == resultCode
}

func IsGenerationError(err error) bool {
	return isError(err, ast.GENERATION_ERROR)
}

func IsKeyNotFoundError(err error) bool {
	return isError(err, ast.KEY_NOT_FOUND_ERROR)
}
