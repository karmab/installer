// +build baremetal

package validation

import (
	"context"
	"fmt"
	pb "github.com/karmab/terraform-provider-kcli/kcli-proto"
	"github.com/openshift/installer/pkg/types/baremetal"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"strings"
	"time"
)

func init() {
	dynamicValidators = append(dynamicValidators, validateInterfaces)
}

// validateInterfaces ensures that any interfaces required by the platform exist on the kcli host.
func validateInterfaces(p *baremetal.Platform, fldPath *field.Path) field.ErrorList {
	errorList := field.ErrorList{}

	findInterface, err := interfaceValidator(p.LibvirtURI)
	if err != nil {
		errorList = append(errorList, field.InternalError(fldPath.Child("libvirtURI"), err))
		return errorList
	}

	if err := findInterface(p.ExternalBridge); err != nil {
		errorList = append(errorList, field.Invalid(fldPath.Child("externalBridge"), p.ExternalBridge, err.Error()))
	}

	if err := findInterface(p.ProvisioningBridge); err != nil {
		errorList = append(errorList, field.Invalid(fldPath.Child("provisioningBridge"), p.ProvisioningBridge, err.Error()))
	}

	return errorList
}

// interfaceValidator fetches the valid interface names from a particular kcli instance, and returns a closure
// to validate if an interface is found among them
func interfaceValidator(Url string) (func(string) error, error) {
	// Connect to libvirt and obtain a list of interface names
	conn, err := grpc.Dial(Url, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "fail to connect to kcli url")
	}
	defer conn.Close()
	k := pb.NewKcliClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	networks, err := k.ListNetworks(ctx, &pb.Empty{})
	if err != nil {
		return nil, errors.Wrap(err, "could not list network interfaces")
	}
	interfaceNames := make([]string, len(networks.Networks))
	for idx, iface := range networks.Networks {
		if err == nil {
			interfaceNames[idx] = iface.Network
		} else {
			return nil, errors.Wrap(err, "could not get interface name from kcli")
		}
	}

	// Return a closure to validate if any particular interface is found among interfaceNames
	return func(interfaceName string) error {
		if len(interfaceNames) == 0 {
			return fmt.Errorf("no interfaces found")
		} else {
			for _, foundInterface := range interfaceNames {
				if foundInterface == interfaceName {
					return nil
				}
			}

			return fmt.Errorf("could not find interface %q, valid interfaces are %s", interfaceName, strings.Join(interfaceNames, ", "))
		}
	}, nil
}
