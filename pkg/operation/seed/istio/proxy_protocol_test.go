// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package istio_test

// This entire test is commented out because istio requires K8S API 1.18
// TODO (mvladev): once we update to 1.18 enable this.

import (
	"context"

	cr "github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/operation/botanist/component"
	. "github.com/gardener/gardener/pkg/operation/seed/istio"

	// . "github.com/gardener/gardener/test/gomega"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// xdsAPI "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	// listenerv2 "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	// "istio.io/api/networking/v1alpha3"
	// "istio.io/api/networking/v1beta1"
	// networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	// networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Proxy protocol", func() {
	const (
		deployNS = "test-chart-namespace"
	)

	var (
		ctx   context.Context
		c     client.Client
		proxy component.DeployWaiter
		// expectedGW *networkingv1beta1.Gateway
		// expectedVS *networkingv1beta1.VirtualService
		// expectedEF *networkingv1alpha3.EnvoyFilter
	)

	BeforeEach(func() {
		ctx = context.TODO()

		s := runtime.NewScheme()
		// Expect(networkingv1beta1.AddToScheme(s)).NotTo(HaveOccurred())
		// Expect(networkingv1alpha3.AddToScheme(s)).NotTo(HaveOccurred())

		c = fake.NewFakeClientWithScheme(s)

		renderer := cr.NewWithServerVersion(&version.Info{})
		ca := kubernetes.NewChartApplier(renderer, kubernetes.NewApplier(c, meta.NewDefaultRESTMapper([]schema.GroupVersion{})))
		Expect(ca).NotTo(BeNil(), "should return chart applier")

		proxy = NewProxyProtocolGateway(deployNS, ca, c, chartsRootPath)

		// expectedGW = &networkingv1beta1.Gateway{
		// 	ObjectMeta: metav1.ObjectMeta{
		// 		Labels: map[string]string{
		// 			"app":   "istio-ingressgateway",
		// 			"istio": "ingressgateway",
		// 		},
		// 	},
		// 	Spec: v1beta1.Gateway{
		// 		Selector: map[string]string{
		// 			"app":   "istio-ingressgateway",
		// 			"istio": "ingressgateway",
		// 		},
		// 		Servers: []*v1beta1.Server{{
		// 			Port: &v1beta1.Port{
		// 				Number:   uint32(8443),
		// 				Name:     "tcp",
		// 				Protocol: "TCP",
		// 			},
		// 			Hosts: []string{"*"},
		// 		}},
		// 	},
		// }

		// expectedVS = &networkingv1beta1.VirtualService{
		// 	ObjectMeta: metav1.ObjectMeta{
		// 		Labels: map[string]string{
		// 			"app":   "istio-ingressgateway",
		// 			"istio": "ingressgateway",
		// 		},
		// 	},
		// 	Spec: v1beta1.VirtualService{
		// 		Hosts:    []string{"blackhole.local"},
		// 		Gateways: []string{"proxy-protocol"},
		// 		ExportTo: []string{"."},
		// 		Tcp: []*v1beta1.TCPRoute{
		// 			{
		// 				Match: []*v1beta1.L4MatchAttributes{{Port: uint32(8443)}},
		// 				Route: []*v1beta1.RouteDestination{
		// 					{Destination: &v1beta1.Destination{
		// 						Host: "localhost",
		// 						Port: &v1beta1.PortSelector{Number: 9999},
		// 					}},
		// 				},
		// 			},
		// 		},
		// 	},
		// }

		// expectedEF = &networkingv1alpha3.EnvoyFilter{
		// 	ObjectMeta: metav1.ObjectMeta{
		// 		Labels: map[string]string{
		// 			"app":   "istio-ingressgateway",
		// 			"istio": "ingressgateway",
		// 		},
		// 	},
		// 	Spec: v1alpha3.EnvoyFilter{
		// 		WorkloadSelector: &v1alpha3.WorkloadSelector{
		// 			Labels: map[string]string{
		// 				"app":   "istio-ingressgateway",
		// 				"istio": "ingressgateway",
		// 			},
		// 		},
		// 		ConfigPatches: []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{{
		// 			ApplyTo: v1alpha3.EnvoyFilter_LISTENER,
		// 			Match: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
		// 				Context: v1alpha3.EnvoyFilter_GATEWAY,
		// 				ObjectTypes: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
		// 					Listener: &v1alpha3.EnvoyFilter_ListenerMatch{
		// 						PortNumber: uint32(8443),
		// 						Name:       "0.0.0.0_8443",
		// 					},
		// 				},
		// 			},
		// 			Patch: &v1alpha3.EnvoyFilter_Patch{
		// 				Operation: v1alpha3.EnvoyFilter_Patch_MERGE,
		// 				Value: messageToStruct(&xdsAPI.Listener{
		// 					ListenerFilters: []*listenerv2.ListenerFilter{
		// 						{Name: "envoy.listener.proxy_protocol"},
		// 					},
		// 				}),
		// 			},
		// 		}},
		// 	},
		// }

	})
	// JustBeforeEach(func() {
	// 	Expect(proxy.Deploy(ctx)).ToNot(HaveOccurred(), "proxy deploy succeeds")
	// })

	It("deploy succeeds", func() {
		Expect(proxy.Deploy(ctx)).ToNot(HaveOccurred(), "proxy deploy succeeds")
	})

	It("destroy succeeds", func() {
		Expect(proxy.Deploy(ctx)).ToNot(HaveOccurred(), "proxy destroy succeeds")
	})

	// It("should deploy blackhole virtual service", func() {
	// 	actualVS := &networkingv1beta1.VirtualService{}
	// 	Expect(c.Get(
	// 		ctx,
	// 		client.ObjectKey{Namespace: deployNS, Name: "proxy-protocol-blackhole"},
	// 		actualVS,
	// 	)).ToNot(HaveOccurred())

	// 	Expect(actualVS).To(DeepDerivativeEqual(expectedVS))
	// })

	// It("should destroy blackhole virtual service", func() {
	// 	Expect(proxy.Destroy(ctx)).ToNot(HaveOccurred(), "destroy succeeds")

	// 	Expect(c.Get(
	// 		ctx,
	// 		client.ObjectKey{Namespace: deployNS, Name: "proxy-protocol-blackhole"},
	// 		&networkingv1beta1.VirtualService{},
	// 	)).To(BeNotFoundError())
	// })

	// It("should deploy envoy filter for proxy protocol", func() {
	// 	actualEF := &networkingv1alpha3.EnvoyFilter{}
	// 	Expect(c.Get(
	// 		ctx,
	// 		client.ObjectKey{Namespace: deployNS, Name: "proxy-protocol"},
	// 		actualEF,
	// 	)).ToNot(HaveOccurred())

	// 	Expect(actualEF).To(DeepDerivativeEqual(expectedEF))
	// })

	// It("should deploy proxy protocol gateway", func() {
	// 	actualGW := &networkingv1beta1.Gateway{}
	// 	Expect(c.Get(
	// 		ctx,
	// 		client.ObjectKey{Namespace: deployNS, Name: "proxy-protocol"},
	// 		actualGW,
	// 	)).ToNot(HaveOccurred())

	// 	Expect(actualGW).To(DeepDerivativeEqual(expectedGW))
	// })
})
