package main

import "io"

func (r *resticWrapper) prepareStdout() (io.WriteCloser, error) {

	if r.profile.Backup != nil && r.profile.Backup.UseStdin {
		if len(r.profile.Backup.StdinCommand) > 0 {
			//return r.prepareCommandStreamSource()
		} else {
			//return r.prepareStdinStreamSource()
		}
	}
	return nil, nil
}
