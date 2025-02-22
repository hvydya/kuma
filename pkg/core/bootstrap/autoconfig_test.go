package bootstrap

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
)

var _ = Describe("Auto configuration", func() {

	It("should autoconfigure xds params", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.DpServer.Port = 1234

		// when
		err := autoconfigure(&cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.BootstrapServer.Params.XdsPort).To(Equal(uint32(1234)))
	})
})
