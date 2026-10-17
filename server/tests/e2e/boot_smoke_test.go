//go:build e2e

package e2e_test

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"jww/tests/utils"
)

// TestCmdServerBoots is the one place that exercises cmd/server's own
// wiring — config loading, database.Migrate running at startup, and
// graceful shutdown on SIGTERM — none of which the in-process journey tests
// above touch (they build kernel.Deps by hand and skip main() entirely).
// Business behavior is already covered by those journeys; this test only
// proves the real entrypoint boots and shuts down cleanly against a real
// database.
func TestCmdServerBoots(t *testing.T) {
	serverRoot := utils.ModuleRoot(t)

	binPath := filepath.Join(t.TempDir(), "jww-server-smoke")
	build := exec.Command("go", "build", "-o", binPath, "./cmd/server")
	build.Dir = serverRoot
	output, err := build.CombinedOutput()
	require.NoErrorf(t, err, "building cmd/server: %s", output)

	port, err := utils.FreePort()
	require.NoError(t, err)
	addr := fmt.Sprintf("http://127.0.0.1:%d", port)

	cmd := exec.Command(binPath)
	cmd.Dir = t.TempDir() // no .env file here; every setting comes from the environment below
	cmd.Env = append(os.Environ(),
		"HTTP_HOST=127.0.0.1",
		fmt.Sprintf("HTTP_PORT=%d", port),
		"DATABASE_URL="+pgPool.DSN,
		"JWT_SECRET_KEY=e2e-boot-smoke-throwaway-secret",
		"LOG_PRETTY=false",
		"OBJECT_STORAGE_S3_BUCKET="+minioStore.Config.Bucket,
		"OBJECT_STORAGE_S3_REGION="+minioStore.Config.Region,
		"OBJECT_STORAGE_S3_ENDPOINT="+minioStore.Config.Endpoint,
		"OBJECT_STORAGE_S3_FORCE_PATH_STYLE=true",
		"OBJECT_STORAGE_S3_ACCESS_KEY_ID="+minioStore.Config.AccessKeyID,
		"OBJECT_STORAGE_S3_SECRET_ACCESS_KEY="+minioStore.Config.SecretAccessKey,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	require.NoError(t, cmd.Start(), "starting cmd/server subprocess")

	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()

	require.Eventually(t, func() bool {
		select {
		case err := <-exited:
			t.Fatalf("server exited early (err=%v); stdout=%s stderr=%s", err, stdout.String(), stderr.String())
			return false
		default:
		}
		resp, err := http.Get(addr + "/api/v1/auth/me")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return true // any HTTP response at all means the server is accepting connections
	}, 15*time.Second, 100*time.Millisecond, "server never became reachable")

	require.NoError(t, cmd.Process.Signal(syscall.SIGTERM))

	select {
	case err := <-exited:
		require.NoErrorf(t, err, "server should exit 0 on graceful shutdown; stdout=%s stderr=%s", stdout.String(), stderr.String())
	case <-time.After(12 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatalf("server did not exit within 12s of SIGTERM (main.go's own shutdown timeout is 10s); stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}
