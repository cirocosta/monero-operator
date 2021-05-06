package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type RunCommand struct {
	Source      string `long:"source" required:"true" description:"source of creds, cfg, etc"`
	Destination string `long:"destination" required:"true" description:"where to save files to"`
}

func init() {
	parser.AddCommand("run",
		"Run Tornetes, the Tor wrapper for Kubernetes",
		"Run Tornetes, the Tor wrapper for Kubernetes",
		&RunCommand{},
	)
}

func (c *RunCommand) Execute(_ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO before starting up, make sure that the files HAVE been populated
	//

	// if err := cpDir(c.Source, c.Destination); err != nil {
	// 	return fmt.Errorf("cp dir: %w", err)
	// }

	if err := os.Chmod(c.Destination, 0700); err != nil {
		return fmt.Errorf("chmod '%s': %w", c.Destination, err)
	}

	cmd, torC, err := startTor(ctx, filepath.Join(c.Destination, "torrc"))
	if err != nil {
		return fmt.Errorf("start tor: %w", err)
	}

	defer cmd.Process.Kill()

	watchC, err := watchFs(ctx, c.Source)
	if err != nil {
		return fmt.Errorf("watchfs: %w", err)
	}

	select {
	case err := <-torC:
		if err != nil {
			return fmt.Errorf("tor err: %w", err)
		}

		return fmt.Errorf("tor shutdown")

	case err := <-watchC:
		if err != nil {
			return fmt.Errorf("watch err: %w", err)
		}
	}

	return nil
}

// watchFs watches fpaths and on any change (or errors), notifies a channel.
//
func watchFs(ctx context.Context, fpaths ...string) (chan error, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("new watcher: %w", err)
	}

	done := make(chan error)
	go func() {
		defer close(done)

		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok { // chan closed
					return
				}

				done <- nil
				return

			case err, ok := <-watcher.Errors:
				if !ok { // chan closed
					return
				}

				done <- fmt.Errorf("watcher err: %w", err)
				return

			case <-ctx.Done():
				if ctx.Err() != nil {
					done <- fmt.Errorf("context err: %w", err)
				}
				return
			}
		}

		close(done)
		watcher.Close()
	}()

	for _, fpath := range fpaths {
		err = watcher.Add(fpath)
		if err != nil {
			return nil, fmt.Errorf("watcher add '%s': %w", fpath, err)
		}
	}

	return done, nil
}

// cpFile copies the contents of a file `src` to `dst` (creating it if it doesn't
// exist, overwriting if it already does) not caring about the permission bits
// (aka, you must chmod/chown later if you care about that).
//
func cpFile(src, dst string) error {
	fmt.Printf("+ cpfile'%s' '%s'\n", src, dst)

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src '%s': %w", src, err)
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst '%s': %w", dst, err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("copy contents from '%s' to '%s': %w", src, dst, err)
	}

	return out.Close()
}
func cpDir(src, dst string) error {
	fmt.Printf("+ cpdir '%s' '%s'\n", src, dst)

	srcinfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return fmt.Errorf("mkdirall: %w", err)
	}

	fds, err := ioutil.ReadDir(src)
	if err != nil {
		return fmt.Errorf("readdir: %w", err)
	}

	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = cpDir(srcfp, dstfp); err != nil {
				return fmt.Errorf("cp dir '%s' to '%s': %w", srcfp, dstfp, err)
			}

			continue
		}

		if err = cpFile(srcfp, dstfp); err != nil {
			return fmt.Errorf("cp file '%s' to '%s': %w", srcfp, dstfp, err)
		}
	}

	return nil
}

func startTor(ctx context.Context, torrc string) (*exec.Cmd, chan error, error) {
	cmd := exec.CommandContext(ctx, "tor", "-f", torrc)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("cmd start: %w", err)
	}

	errC := make(chan error)
	go func() {
		errC <- cmd.Wait()
	}()

	return cmd, errC, nil
}
