package servicecatalogaddons

import (
	"context"

	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/resource"

	"github.com/golang/glog"
	api "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalogaddons/listener"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlerror"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=serviceBindingUsageOperations -output=automock -outpkg=automock -case=underscore
type serviceBindingUsageOperations interface {
	Create(namespace string, sb *api.ServiceBindingUsage) (*api.ServiceBindingUsage, error)
	Delete(namespace string, name string) error
	Find(namespace string, name string) (*api.ServiceBindingUsage, error)
	ListForServiceInstance(namespace string, instanceName string) ([]*api.ServiceBindingUsage, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

type serviceBindingUsageResolver struct {
	operations serviceBindingUsageOperations
	converter  serviceBindingUsageConverter
}

func newServiceBindingUsageResolver(op serviceBindingUsageOperations) *serviceBindingUsageResolver {
	return &serviceBindingUsageResolver{
		operations: op,
		converter:  newBindingUsageConverter(),
	}
}

func (r *serviceBindingUsageResolver) CreateServiceBindingUsageMutation(ctx context.Context, input *gqlschema.CreateServiceBindingUsageInput) (*gqlschema.ServiceBindingUsage, error) {
	inBindingUsage, err := r.converter.InputToK8s(input)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s from input [%+v]", pretty.ServiceBindingUsage, input))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage)
	}
	bu, err := r.operations.Create(input.Namespace, inBindingUsage)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s from input [%v]", pretty.ServiceBindingUsage, input))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(*input.Name), gqlerror.WithNamespace(input.Namespace))
	}

	out, err := r.converter.ToGQL(bu)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceBindingUsage))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(*input.Name), gqlerror.WithNamespace(input.Namespace))
	}

	return out, nil
}

func (r *serviceBindingUsageResolver) DeleteServiceBindingUsageMutation(ctx context.Context, serviceBindingUsageName, namespace string) (*gqlschema.DeleteServiceBindingUsageOutput, error) {
	err := r.operations.Delete(namespace, serviceBindingUsageName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s with name `%s` from namespace `%s`", pretty.ServiceBindingUsage, serviceBindingUsageName, namespace))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(serviceBindingUsageName), gqlerror.WithNamespace(namespace))
	}

	return &gqlschema.DeleteServiceBindingUsageOutput{
		Namespace: namespace,
		Name:      serviceBindingUsageName,
	}, nil
}

func (r *serviceBindingUsageResolver) ServiceBindingUsageQuery(ctx context.Context, name, namespace string) (*gqlschema.ServiceBindingUsage, error) {
	usage, err := r.operations.Find(namespace, name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting single %s [name: %s, namespace: %s]", pretty.ServiceBindingUsage, name, namespace))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	out, err := r.converter.ToGQL(usage)
	if err != nil {
		glog.Error(
			errors.Wrapf(err,
				"while getting single %s [name: %s, namespace: %s]: while converting %s to QL representation", pretty.ServiceBindingUsage,
				name, namespace, pretty.ServiceBindingUsage))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsage, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	return out, nil
}

func (r *serviceBindingUsageResolver) ServiceBindingUsagesOfInstanceQuery(ctx context.Context, instanceName, namespace string) ([]gqlschema.ServiceBindingUsage, error) {
	usages, err := r.operations.ListForServiceInstance(namespace, instanceName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s of instance [namespace: %s, instance: %s]", pretty.ServiceBindingUsages, namespace, instanceName))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsages, gqlerror.WithNamespace(namespace), gqlerror.WithCustomArgument("instanceName", instanceName))
	}
	out, err := r.converter.ToGQLs(usages)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s of instance [namespace: %s, instance: %s]", pretty.ServiceBindingUsages, namespace, instanceName))
		return nil, gqlerror.New(err, pretty.ServiceBindingUsages, gqlerror.WithNamespace(namespace), gqlerror.WithCustomArgument("instanceName", instanceName))
	}
	return out, nil
}

func (r *serviceBindingUsageResolver) ServiceBindingUsageEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ServiceBindingUsageEvent, error) {
	channel := make(chan gqlschema.ServiceBindingUsageEvent, 1)
	filter := func(bindingUsage *api.ServiceBindingUsage) bool {
		return bindingUsage != nil && bindingUsage.Namespace == namespace
	}

	bindingUsageListener := listener.NewBindingUsage(channel, filter, &r.converter)

	r.operations.Subscribe(bindingUsageListener)
	go func() {
		defer close(channel)
		defer r.operations.Unsubscribe(bindingUsageListener)
		<-ctx.Done()
	}()

	return channel, nil
}
