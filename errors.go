package optidash

import (
    "errors"
    "fmt"
)

// OptidashError is an error returned by the Optidash API.
type OptidashError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

// Error implements the error interface.
func (c *OptidashError) Error() string {
    return fmt.Sprintf("optidash: [%d] %s", c.Code, c.Message)
}

// In-code validation and other frequent error cases.
var (
    ErrInvalidSourceType = errors.New("optidash: Invalid request source type")
    ErrBinaryWebhook     = errors.New("optidash: Webhooks are not supported when using binary responses")
    ErrBinaryStorage     = errors.New("optidash: External storage is not supported when using binary responses")
    ErrNoSuccess         = errors.New("optidash: Success is missing in the response")
)
