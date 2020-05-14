package optidash

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "io/ioutil"
    "mime/multipart"
    "net/http"
    "os"
    "github.com/valyala/fastjson"
)

// Request is a generator that allows creation of Optidash API requests.
type Request struct {
    client *Client
    http   *http.Client

    source   source
    reader   io.Reader
    location string
    context  context.Context

    optimize  P
    flip      P
    resize    P
    scale     P
    crop      P
    watermark P
    mask      P
    stylize   P
    adjust    P
    auto      P
    border    P
    padding   P
    store     P
    output    P
    webhook   P
    response  P
    cdn       P
}

// There are three sources of data:
// - a Reader source, when user passes an io.Reader to Upload()
// - a Path soruce, when user passes a path string to Upload()
// - a Fetch source, when user passes an URL to Fetch()
type source int

const (
    readerSource source = iota
    pathSource
    fetchSource
)

// P is a short name for all hashes passed to the Optidash API.
type P map[string]interface{}

// HTTPClient replaces the client used to execute the request.
func (r *Request) HTTPClient(client *http.Client) *Request {
    r.http = client
    return r
}

// Optimize adds an image optimization step to the transformation flow.
// Check out Optidash docs for more details.
func (r *Request) Optimize(data P) *Request {
    r.optimize = data
    return r
}

// Flip adds an image flipping step to the transformation flow.
// Check out Optidash docs for more details.
func (r *Request) Flip(data P) *Request {
    r.flip = data
    return r
}

// Resize adds an image resizing step to the transformation flow.
// Check out Optidash docs for more details.
func (r *Request) Resize(data P) *Request {
    r.resize = data
    return r
}

// Scale adds an image scaling step to the transformation flow.
// Check out Optidash docs for more details.
func (r *Request) Scale(data P) *Request {
    r.scale = data
    return r
}

// Crop adds an image cropping step to the transformation flow.
// Check out Optidash docs for more details.
func (r *Request) Crop(data P) *Request {
    r.crop = data
    return r
}

// Watermark adds a watermark application to the transformation flow.
// Check out Optidash docs for more details.
func (r *Request) Watermark(data P) *Request {
    r.watermark = data
    return r
}

// Mask adds application of an elliptical mask to the transformation flow.
// Check out Optidash docs for more details.
func (r *Request) Mask(data P) *Request {
    r.mask = data
    return r
}

// Stylize adds filter application to the transformation flow.
// Check out Optidash docs for more details.
func (r *Request) Stylize(data P) *Request {
    r.stylize = data
    return r
}

// Adjust adds an visual parameters adjustment to the transformation flow.
// Check out Optidash docs for more details.
func (r *Request) Adjust(data P) *Request {
    r.adjust = data
    return r
}

// Auto adds an automatic image enhancement step to the transformation flow.
// Check out Optidash docs for more details.
func (r *Request) Auto(data P) *Request {
    r.auto = data
    return r
}

// Border adds adding a border to the image to the transformation flow.
// Check out Optidash docs for more details.
func (r *Request) Border(data P) *Request {
    r.border = data
    return r
}

// Padding adds an image padding step to the transformation flow.
// Check out Optidash docs for more details.
func (r *Request) Padding(data P) *Request {
    r.padding = data
    return r
}

// Store specifies where the image should be stored after transformations.
// Check out Optidash docs for more details.
func (r *Request) Store(data P) *Request {
    r.store = data
    return r
}

// Output sets the output format and encoding.
// Check out Optidash docs for more details.
func (r *Request) Output(data P) *Request {
    r.output = data
    return r
}

// Webhook sets a webhook as a response delivery method.
// Check out Optidash docs for more details.
func (r *Request) Webhook(data P) *Request {
    r.webhook = data
    return r
}

// CDN configures CDN settings of the platform.
// Check out Optidash docs for more details.
func (r *Request) CDN(data P) *Request {
    r.cdn = data
    return r
}

// Context sets the context of the HTTP request.
func (r *Request) Context(ctx context.Context) *Request {
    r.context = ctx
    return r
}

// Internal execution function of HTTP requests.
func (r *Request) execute() (*http.Response, error) {
    // First use this hack to create a map with params
    params := map[string]interface{}{}

    if r.resize != nil {
        params["resize"] = r.resize
    }
    if r.scale != nil {
        params["scale"] = r.scale
    }
    if r.crop != nil {
        params["crop"] = r.crop
    }
    if r.watermark != nil {
        params["watermark"] = r.watermark
    }
    if r.mask != nil {
        params["mask"] = r.mask
    }
    if r.stylize != nil {
        params["stylize"] = r.stylize
    }
    if r.adjust != nil {
        params["adjust"] = r.adjust
    }
    if r.auto != nil {
        params["auto"] = r.auto
    }
    if r.border != nil {
        params["border"] = r.border
    }
    if r.padding != nil {
        params["padding"] = r.padding
    }
    if r.store != nil {
        params["store"] = r.store
    }
    if r.output != nil {
        params["output"] = r.output
    }
    if r.webhook != nil {
        params["webhook"] = r.webhook
    }
    if r.response != nil {
        params["response"] = r.response
    }
    if r.cdn != nil {
        params["cdn"] = r.cdn
    }
    if r.source == fetchSource {
        params["url"] = r.location
    }

    // Then encode it
    pb, err := json.Marshal(params)
    if err != nil {
        return nil, err
    }

    // Decide on URL, Content-Type header and body dpeending on the source of the image.
    var (
        url         string
        contentType string
        body        io.Reader
    )
    if r.source == fetchSource {
        // .Fetch(url) is straightforward
        url = apiURL + "/fetch"
        body = bytes.NewReader(pb)
        contentType = "application/json"
    } else if r.source == readerSource || r.source == pathSource {
        // .Upload(<any>) has two cases
        url = apiURL + "/upload"

        var file io.Reader
        if r.source == readerSource {
            // .Upload(<io.Reader>) simply passes the stream further down the pipeline.
            file = r.reader
        } else if r.source == pathSource {
            // .Upload(path) requires the file to be opened.
            fs, err := os.OpenFile(r.location, os.O_RDONLY, 0600)
            if err != nil {
                return nil, err
            }
            defer fs.Close() // close it when the function ends
            file = fs
        }

        // Prepare a buffer for the multipart body writer
        buf := &bytes.Buffer{}
        body = buf
        writer := multipart.NewWriter(buf)

        // Create the file upload part
        part, err := writer.CreateFormFile("file", "")
        if err != nil {
            return nil, err
        }
        if _, err := io.Copy(part, file); err != nil {
            return nil, err
        }

        // Insert the JSON data into a "data" field
        if err := writer.WriteField("data", string(pb)); err != nil {
            return nil, err
        }

        // End the writing
        if err := writer.Close(); err != nil {
            return nil, err
        }

        // Set the content type accordingly
        contentType = writer.FormDataContentType()
    } else {
        // shouldn't happen
        return nil, ErrInvalidSourceType
    }

    // Create a new HTTP request using previously computed data.
    request, err := http.NewRequest("POST", url, body)
    if err != nil {
        return nil, err
    }

    // Apply headers - Content-Type, Binary and Authorization
    request.Header.Set("Content-Type", contentType)

    // Set the context of the request
    if r.context != nil {
        request = request.WithContext(r.context)
    }

    if r.response != nil {
        if mode, ok := r.response["mode"]; ok && mode == "binary" {
            request.Header.Set("X-Optidash-Binary", "1")
        }
    }

    request.SetBasicAuth(r.client.Key, "")

    // Run the request using the passed Client
    return r.http.Do(request)
}

// ToJSON runs the request and returns a fastjson.Value with a result from the API.
func (r *Request) ToJSON() (*fastjson.Value, error) {
    // Run the request
    resp, err := r.execute()
    if err != nil {
        return nil, err
    }

    // Decode the JSON body into a *fastjson.Value
    var parser fastjson.Parser
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    result, err := parser.ParseBytes(body)
    if err != nil {
        return nil, err
    }

    // Ensuring that the code won't panic, safely generate an API error struct.
    if result != nil {
        value := result.Get("success")
        if value == nil {
            return nil, ErrNoSuccess
        }
        success, err := value.Bool()
        if err != nil {
            return nil, ErrNoSuccess
        }

        if !success {
            err := &OptidashError{}

            // Populate error's code field
            codeValue := result.Get("code")
            if codeValue == nil {
                return nil, ErrNoSuccess
            }
            if code, err2 := codeValue.Int(); err2 == nil {
                err.Code = code
            }

            // And the error message
            messageValue := result.Get("message")
            if messageValue == nil {
                return nil, ErrNoSuccess
            }
            if message, err2 := messageValue.StringBytes(); err2 == nil {
                err.Message = string(message)
            }

            return nil, err
        }
    }

    // The query succeeded!
    return result, nil
}

// ToReader executes the request, waits for the result and returns a set of 3 variables:
//  - meta map containing the result you would normally get in the body
//  - io.ReadCloser containing the resulting file
//  - error that should be nil if everything succeeded
// Due to the fact that ToReader performs a binary request, using Webhook
// and Store is forbidden.
func (r *Request) ToReader() (*fastjson.Value, io.ReadCloser, error) {
    if r.webhook != nil {
        return nil, nil, ErrBinaryWebhook
    }

    if r.store != nil {
        return nil, nil, ErrBinaryStorage
    }

    // Gets embedded into the request
    r.response = P{
        "mode": "binary",
    }

    // Execute the request
    resp, err := r.execute()

    // Clean up the body if execution fails
    var succeeded bool
    defer func() {
        if !succeeded && resp != nil && resp.Body != nil {
            resp.Body.Close()
        }
    }()

    if err != nil {
        return nil, nil, err
    }

    // Try to read the meta object
    var meta *fastjson.Value
    if sm := resp.Header.Get("X-Optidash-Meta"); sm != "" {
        // Decode the JSON body into a *fastjson.Value
        var parser fastjson.Parser
        meta, err = parser.Parse(sm)
        if err != nil {
            return nil, nil, err
        }
    }

    // Check for success and generate an error if something went wrong.
    if meta != nil {
        successValue := meta.Get("success")
        if successValue == nil {
            return nil, nil, ErrNoSuccess
        }
        success, err := successValue.Bool()
        if err != nil {
            return nil, nil, ErrNoSuccess
        }

        if !success {
            err := &OptidashError{}

            // Populate error's code field
            codeValue := meta.Get("code")
            if codeValue == nil {
                return nil, nil, ErrNoSuccess
            }
            if code, err2 := codeValue.Int(); err2 == nil {
                err.Code = code
            }

            // And the error message
            messageValue := meta.Get("message")
            if messageValue == nil {
                return nil, nil, ErrNoSuccess
            }
            if message, err2 := messageValue.StringBytes(); err2 == nil {
                err.Message = string(message)
            }

            return nil, nil, err
        }
    }

    // Make sure that the defer function won't close the body
    succeeded = true

    // Everything succeeded.
    return meta, resp.Body, nil
}

// ToFile executes a request, waits for the results and saves them into a file
// created on given path. If a file does not exist, it creates one with the
// passed `perm` file permissions.
// It returns 2 variables:
//  - meta map containing the result information
//  - error that should be nil if everything succeeded
// Due to the fact that ToFile performs a binary request, using Webhook
// and Store is forbidden.
func (r *Request) ToFile(input string, perm os.FileMode) (*fastjson.Value, error) {
    // First part of the functionality is the same as ToReader
    meta, reader, err := r.ToReader()
    if err != nil {
        return nil, err
    }

    // Create the file, truncate it if it exists.
    file, err := os.OpenFile(input, os.O_CREATE|os.O_RDWR|os.O_TRUNC, perm)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Copy the stream contents into the file
    if _, err := io.Copy(file, reader); err != nil {
        return nil, err
    }

    // Return the meta array, effectively closing the file.
    return meta, nil
}

// CopyTo executes a request, waits for the results and copies it into the
// passed io.Writer.
// It returns 2 variables:
//  - meta map containing the result information
//  - error that should be nil if everything succeeded
// Due to the fact that CopyTo performs a binary request, using Webhook
// and Store is forbidden.
func (r *Request) CopyTo(input io.Writer) (*fastjson.Value, error) {
    // First part of the functionality is the same as ToReader
    meta, reader, err := r.ToReader()
    if err != nil {
        return nil, err
    }

    // Copy the stream contents into the file
    if _, err := io.Copy(input, reader); err != nil {
        return nil, err
    }

    // Return the meta array, effectively closing the file.
    return meta, nil
}
