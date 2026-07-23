package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/creativeprojects/clog"
)

// serveMetricsCommand exposes the active profile's prometheus-save-to-file
// together with Go runtime metrics on /metrics. It is intended for two
// scenarios:
//
//  1. As a long-running sidecar next to `resticprofile schedule`, sharing
//     the same prometheus-save-to-file so Prometheus can scrape the latest
//     backup status without polling per-profile sidecars.
//  2. As a stand-alone helper to debug what's actually written to the
//     textfile by `resticprofile backup`.
//
// The command is registered as experimental/hidden; the recommended path is
// `resticprofile --metrics-port <port> schedule --all`, which keeps the
// process alive and blocks on SIGINT/SIGTERM.
func serveMetricsCommand(ctx commandContext) error {
	port := ctx.flags.metricsPort
	if port <= 0 {
		return fmt.Errorf("missing metrics port: set --metrics-port or RESTICPROFILE_METRICS_PORT")
	}

	profileName := ctx.request.profile
	if profileName == "" {
		return fmt.Errorf("missing profile: pass --name <profile> or use the <profile>.%s form", ctx.request.command)
	}
	profile, err := ctx.config.GetProfile(profileName)
	if err != nil {
		return fmt.Errorf("loading profile %q: %w", profileName, err)
	}
	if profile.PrometheusSaveToFile == "" {
		return fmt.Errorf("profile %q has no prometheus-save-to-file configured", profileName)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	defer signal.Stop(quit)

	clog.Infof("serving metrics for profile %q", profileName)
	return newMetricsServer(port, profile.PrometheusSaveToFile).run(quit)
}
