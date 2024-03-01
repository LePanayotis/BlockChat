package blockchat

import(
	"os"
	"log/slog"
)

var logger *slog.Logger = slog.New(slog.NewJSONHandler(os.Stderr,nil))

type TeeWriter struct {
    stdout *os.File
    file   *os.File
}

func (t *TeeWriter) Write(p []byte) (n int, err error) {
    n, err = t.stdout.Write(p)
    if err != nil {
        return n, err
    }
    n, err = t.file.Write(p)
    return n, err
}