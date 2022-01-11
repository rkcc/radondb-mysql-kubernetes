/*
Copyright 2021 RadonDB.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
)

var _ = Describe("MySQL Cluster E2E Tests", Label("Cluster"), func() {
	var (
		f          *framework.Framework
		cluster    *apiv1alpha1.MysqlCluster
		clusterKey *types.NamespacedName
		two        = int32(2)
		three      = int32(3)
		five       = int32(5)
	)

	BeforeEach(func() {
		// Singleton
		if f == nil {
			By("Init framework")
			f = &framework.Framework{
				BaseName: "mysqlcluster-e2e",
				Log:      framework.Log,
			}
			f.BeforeEach()
		}
		Expect(f).ShouldNot(BeNil(), "failed to init framework")

		clusterKey = &types.NamespacedName{Namespace: framework.RadondbMysqlE2eNamespace}
		if clusterName := getExistCluster(f); clusterName != "" {
			By("Reuse the exist cluster")
			clusterKey.Name = clusterName
		} else {
			By("Creating a new testing cluster")
			clusterKey.Name = f.InitAClusterForTesting()
		}

		By("Testing the cluster readiness")
		cluster = &apiv1alpha1.MysqlCluster{}
		Expect(f.Client.Get(context.TODO(), *clusterKey, cluster)).To(Succeed(), "failed to get cluster %s", cluster.Name)
		f.WaitClusterReadiness(cluster)
	})

	// Run the full scale in/out test with label filter: Scale.
	// Run only scale out(2 -> 3 -> 5): Scale out.
	// Run only scale in(5 -> 3 -> 2): Scale in.
	When("Test cluster scale in/out", Label("Scale"), Ordered, func() {
		Context("Scale out", Label("Scale out"), Ordered, func() {
			// Guarantee the initial replicas is 2.
			BeforeAll(func() {
				cluster.Spec.Replicas = &two
				Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
				f.WaitClusterReadiness(cluster)
			})

			Specify("Replicas: 2 -> 3", func() {
				cluster.Spec.Replicas = &three
				Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())

				By("Wait scale out finished")
				fmt.Println("Time length: ", f.WaitClusterReadiness(cluster))
			})

			Specify("Replicas: 3 -> 5", func() {
				cluster.Spec.Replicas = &five
				Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())

				By("Wait scale out finished")
				fmt.Println("Time length: ", f.WaitClusterReadiness(cluster))
			})
		})

		Context("Scale in", Label("Scale In"), Ordered, func() {
			// Guarantee the initial replicas is 5.
			BeforeAll(func() {
				cluster.Spec.Replicas = &five
				Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
				f.WaitClusterReadiness(cluster)
			})

			Specify("Replicas: 5 -> 3", func() {
				cluster.Spec.Replicas = &three
				Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())

				By("Wait scale in finished")
				fmt.Println("Time length: ", f.WaitClusterReadiness(cluster))
			})
			Specify("Replicas: 3 -> 2", func() {
				cluster.Spec.Replicas = &two
				Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())

				By("Wait scale in finished")
				fmt.Println("Time length: ", f.WaitClusterReadiness(cluster))
			})
		})
	})
})

func getExistCluster(f *framework.Framework) string {
	existClusters := &apiv1alpha1.MysqlClusterList{}
	Expect(f.Client.List(context.TODO(), existClusters, &client.ListOptions{
		Namespace: framework.RadondbMysqlE2eNamespace,
	})).To(Succeed(), "failed to list clusters")

	if len(existClusters.Items) > 0 {
		return existClusters.Items[0].Name
	}
	return ""
}
