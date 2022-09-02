package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	"github.com/thanos-io/thanos/pkg/discovery/dns"

	"github.com/prometheus/common/version"
	"github.com/thanos-io/thanos/pkg/extkingpin"
	"github.com/thanos-io/thanos/pkg/logging"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := extkingpin.NewApp(kingpin.New(filepath.Base(os.Args[0]), "nslookup test utility.").Version(version.Print("thanos")))
	address := app.Flag("address", "Address to lookup.").Required().String()
	interval := app.Flag("interval", "Interval between successive lookups.").Default("1s").Duration()
	logFormat := app.Flag("log.format", "Log format to use. Possible options: logfmt or json.").Default(logging.LogFormatLogfmt).Enum(logging.LogFormatLogfmt, logging.LogFormatJSON)
	logLevel := app.Flag("log.level", "Log filtering level.").Default("info").Enum("error", "warn", "info", "debug")
	resolverName := app.Flag("resolver.name", "Resolver implementation.").Default("miekgdns").String()

	app.Parse()

	logger := logging.NewLogger(*logLevel, *logFormat, *resolverName)

	var resolver dns.Resolver

	switch *resolverName {
	case "miekgdns":
		resolver = dns.NewResolver(dns.MiekgdnsResolverType.ToResolver(logger), logger)
	default:
		resolver = dns.NewResolver(dns.GolangResolverType.ToResolver(logger), logger)
	}

	for {
		if values, err := resolver.Resolve(context.TODO(), *address, dns.A); err != nil {
			level.Error(logger).Log("err", errors.Wrapf(err, "lookup failed: %v", err))
		} else {
			level.Info(logger).Log("records", fmt.Sprintf("%v", values))
		}
		time.Sleep(*interval)
	}
}
