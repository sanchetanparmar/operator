/*
 * This file is part of MinIO Operator
 * Copyright (C) 2020, MinIO, Inc.
 *
 * This code is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License, version 3,
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 */

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/minio/kubectl-minio/cmd/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	operatorv1 "github.com/minio/operator/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
)

const (
	tenantDeleteDesc = `
'delete' command deletes a MinIO tenant`
	tenantDeleteExample = `  kubectl minio tenant delete --name tenant1 --namespace tenant1-ns`
)

type tenantDeleteCmd struct {
	out    io.Writer
	errOut io.Writer
	name   string
	ns     string
}

func newTenantDeleteCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &tenantDeleteCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a MinIO tenant",
		Long:    deleteDesc,
		Example: deleteExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(args)
		},
	}

	f := cmd.Flags()
	f.StringVar(&c.name, "name", "", "name of the MinIO tenant to delete")
	f.StringVarP(&c.ns, "namespace", "n", helpers.DefaultNamespace, "namespace scope for this request")
	return cmd
}

func (d *tenantDeleteCmd) validate() error {
	if d.name == "" {
		return errors.New("--name flag is required for tenant deletion")
	}
	return nil
}

// run initializes local config and installs MinIO Operator to Kubernetes cluster.
func (d *tenantDeleteCmd) run(args []string) error {
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	kclient, err := helpers.GetKubeClient()
	if err != nil {
		return err
	}
	return deleteTenant(oclient, kclient, d)
}

func deleteTenant(client *operatorv1.Clientset, kclient *kubernetes.Clientset, d *tenantDeleteCmd) error {
	tenant, err := client.MinioV1().Tenants(d.ns).Get(context.Background(), d.name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if err := client.MinioV1().Tenants(d.ns).Delete(context.Background(), d.name, v1.DeleteOptions{}); err != nil {
		return err
	}
	if err := kclient.CoreV1().Secrets(d.ns).Delete(context.Background(), tenant.Spec.CredsSecret.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}
	if err := kclient.CoreV1().Secrets(d.ns).Delete(context.Background(), tenant.Spec.Console.ConsoleSecret.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	fmt.Printf("Deleting MinIO Tenant %s\n", tenant.ObjectMeta.Name)
	fmt.Printf("Deleting MinIO Tenant Credentials Secret %s\n", tenant.Spec.CredsSecret.Name)
	fmt.Printf("Deleting MinIO Tenant Console Secret %s\n", tenant.Spec.Console.ConsoleSecret.Name)
	return nil
}
