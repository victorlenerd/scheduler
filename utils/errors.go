package utils

type GenericError struct {
	Message string
	Type    int
}

func (g GenericError) Error() string {
	panic("implement me")
}

func HTTPGenericError(httpStatus int, errorMessage string) *GenericError {
	return &GenericError{
		Type:    httpStatus,
		Message: errorMessage,
	}
}

func CheckErr(e error) {
	if e != nil {
		panic(e)
	}
}
