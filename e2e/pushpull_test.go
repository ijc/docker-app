package e2e

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/app/internal"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	"gotest.tools/fs"
	"gotest.tools/icmd"
)

func TestPushInstall(t *testing.T) {
	registryPort := findAvailablePort()
	tmpDir := fs.NewDir(t, t.Name())
	defer tmpDir.Remove()

	cmd := icmd.Cmd{
		Env: append(os.Environ(),
			fmt.Sprintf("DUFFLE_HOME=%s", tmpDir.Path()),
			fmt.Sprintf("DOCKER_CONFIG=%s", tmpDir.Path()),
			"DOCKER_TARGET_CONTEXT=swarm-target-context",
		),
	}

	// we have a difficult constraint here:
	// - the registry must be reachable from the client side (for cnab-to-oci, which does not use the docker daemon to access the registry)
	// - the registry must be reachable from the dind daemon on the same address/port
	// Solution found is: fix the port of the registry to be the same internally and externally (fixed at 5000, could use something random)
	// run the dind container in the same network namespace: this way 127.0.0.1:5000 both resolves to the registry from the client and from dind

	registry := NewContainer("registry:2", registryPort)
	registry.Start(t, "-e", "REGISTRY_VALIDATION_MANIFESTS_URLS_ALLOW=[^http]",
		"-e", fmt.Sprintf("REGISTRY_HTTP_ADDR=0.0.0.0:%d", registryPort),
		"--expose", "2375",
		"-p", fmt.Sprintf("%d:%d", registryPort, registryPort),
		"-p", "2375")
	defer registry.Stop(t)

	ref := registry.GetAddress(t) + "/test/push-pull"
	cmd.Command = []string{dockerApp, "push", "-t", ref, "--insecure-registries=" + registry.GetAddress(t), filepath.Join("testdata", "push-pull", "push-pull.dockerapp")}
	icmd.RunCmd(cmd).Assert(t, icmd.Success)

	swarm := NewContainer("docker:18.09-dind", 2375, "--insecure-registry", registry.GetAddress(t))
	swarm.StartWithContainerNetwork(t, registry)
	defer swarm.Stop(t)
	// The dind doesn't have the cnab-app-base image so we save it in order to load it later
	icmd.RunCommand(dockerCli, "save", fmt.Sprintf("docker/cnab-app-base:%s", internal.Version), "-o", tmpDir.Join("cnab-app-base.tar.gz")).Assert(t, icmd.Success)

	// We  need two contexts:
	// - one for `docker` so that it connects to the dind swarm created before
	// - the target context for the invocation image to install within the swarm
	cmd.Command = []string{dockerCli, "context", "create", "swarm-context", "--docker", fmt.Sprintf(`"host=tcp://%s"`, swarm.GetAddress(t)), "--default-stack-orchestrator", "swarm"}
	icmd.RunCmd(cmd).Assert(t, icmd.Success)

	// When creating a context on a Windows host we cannot use
	// the unix socket but it's needed inside the invocation image.
	// The workaround is to create a context with an empty host.
	// This host will default to the unix socket inside the
	// invocation image
	cmd.Command = []string{dockerCli, "context", "create", "swarm-target-context", "--docker", "host=", "--default-stack-orchestrator", "swarm"}
	icmd.RunCmd(cmd).Assert(t, icmd.Success)

	// Initialize the swarm
	cmd.Env = append(cmd.Env, "DOCKER_CONTEXT=swarm-context")
	cmd.Command = []string{dockerCli, "swarm", "init"}
	icmd.RunCmd(cmd).Assert(t, icmd.Success)
	// Load the needed base cnab image into the swarm docker engine
	cmd.Command = []string{dockerCli, "load", "-i", tmpDir.Join("cnab-app-base.tar.gz")}
	icmd.RunCmd(cmd).Assert(t, icmd.Success)

	cmd.Command = []string{dockerApp, "install", "--insecure-registries=" + registry.GetAddress(t), ref, "--name", t.Name()}
	icmd.RunCmd(cmd).Assert(t, icmd.Success)
	cmd.Command = []string{dockerCli, "service", "ls"}
	assert.Check(t, cmp.Contains(icmd.RunCmd(cmd).Assert(t, icmd.Success).Combined(), ref))
}

func findAvailablePort() int {
	rand.Seed(time.Now().UnixNano())
	for {
		candidate := (rand.Int() % 2000) + 5000
		if isPortAvailable(candidate) {
			return candidate
		}
	}
}

func isPortAvailable(port int) bool {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	defer l.Close()
	return true
}
