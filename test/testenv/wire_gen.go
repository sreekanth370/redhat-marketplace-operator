// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package testenv

import (
	"github.com/google/wire"
	"github.com/redhat-marketplace/redhat-marketplace-operator/pkg/controller"
	"github.com/redhat-marketplace/redhat-marketplace-operator/pkg/utils/reconcileutils"
)

// Injectors from wire.go:

func initializeLocalSchemes() controller.LocalSchemes {
	opsSrcSchemeDefinition := controller.ProvideOpsSrcScheme()
	monitoringSchemeDefinition := controller.ProvideMonitoringScheme()
	olmV1SchemeDefinition := controller.ProvideOLMV1Scheme()
	olmV1Alpha1SchemeDefinition := controller.ProvideOLMV1Alpha1Scheme()
	localSchemes := controller.ProvideLocalSchemes(opsSrcSchemeDefinition, monitoringSchemeDefinition, olmV1SchemeDefinition, olmV1Alpha1SchemeDefinition)
	return localSchemes
}

func initializeControllers() controller.ControllerList {
	marketplaceController := controller.ProvideMarketplaceController()
	defaultCommandRunnerProvider := reconcileutils.ProvideDefaultCommandRunnerProvider()
	meterbaseController := controller.ProvideMeterbaseController(defaultCommandRunnerProvider)
	meterDefinitionController := controller.ProvideMeterDefinitionController(defaultCommandRunnerProvider)
	razeeDeployController := controller.ProvideRazeeDeployController()
	olmSubscriptionController := controller.ProvideOlmSubscriptionController()
	controllerList := controller.ProvideControllerList(marketplaceController, meterbaseController, meterDefinitionController, razeeDeployController, olmSubscriptionController)
	return controllerList
}

// wire.go:

var TestControllerSet = wire.NewSet(controller.ControllerSet, controller.ProvideControllerFlagSet, controller.SchemeDefinitions, reconcileutils.ProvideDefaultCommandRunnerProvider, wire.Bind(new(reconcileutils.ClientCommandRunnerProvider), new(*reconcileutils.DefaultCommandRunnerProvider)))