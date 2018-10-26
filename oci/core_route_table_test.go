// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/oracle/oci-go-sdk/common"
	oci_core "github.com/oracle/oci-go-sdk/core"
)

var (
	RouteTableRequiredOnlyResource = RouteTableResourceDependencies +
		generateResourceFromRepresentationMap("oci_core_route_table", "test_route_table", Required, Create, routeTableRepresentation)

	routeTableDataSourceRepresentation = map[string]interface{}{
		"compartment_id": Representation{repType: Required, create: `${var.compartment_id}`},
		"vcn_id":         Representation{repType: Required, create: `${oci_core_vcn.test_vcn.id}`},
		"display_name":   Representation{repType: Optional, create: `MyRouteTable`, update: `displayName2`},
		"state":          Representation{repType: Optional, create: `AVAILABLE`},
		"filter":         RepresentationGroup{Required, routeTableDataSourceFilterRepresentation}}
	routeTableDataSourceFilterRepresentation = map[string]interface{}{
		"name":   Representation{repType: Required, create: `id`},
		"values": Representation{repType: Required, create: []string{`${oci_core_route_table.test_route_table.id}`}},
	}

	routeTableRepresentation = map[string]interface{}{
		"compartment_id": Representation{repType: Required, create: `${var.compartment_id}`},
		"route_rules":    RepresentationGroup{Required, routeTableRouteRulesRepresentation},
		"vcn_id":         Representation{repType: Required, create: `${oci_core_vcn.test_vcn.id}`},
		"defined_tags":   Representation{repType: Optional, create: `${map("${oci_identity_tag_namespace.tag-namespace1.name}.${oci_identity_tag.tag1.name}", "value")}`, update: `${map("${oci_identity_tag_namespace.tag-namespace1.name}.${oci_identity_tag.tag1.name}", "updatedValue")}`},
		"display_name":   Representation{repType: Optional, create: `MyRouteTable`, update: `displayName2`},
		"freeform_tags":  Representation{repType: Optional, create: map[string]string{"Department": "Finance"}, update: map[string]string{"Department": "Accounting"}},
	}
	routeTableRouteRulesRepresentation = map[string]interface{}{
		"network_entity_id": Representation{repType: Required, create: `${oci_core_internet_gateway.test_network_entity.id}`},
		"cidr_block":        Representation{repType: Required, create: `0.0.0.0/0`, update: `10.0.0.0/8`},
		"destination":       Representation{repType: Optional, create: `0.0.0.0/0`, update: `10.0.0.0/8`},
		"destination_type":  Representation{repType: Optional, create: `CIDR_BLOCK`},
	}
	routeTableRouteRulesRepresentationWithServiceCidr = map[string]interface{}{
		"network_entity_id": Representation{repType: Required, create: `${oci_core_service_gateway.test_service_gateway.id}`},
		"destination":       Representation{repType: Required, create: `${lookup(data.oci_core_services.test_services.services[0], "cidr_block")}`},
		"destination_type":  Representation{repType: Required, create: `SERVICE_CIDR_BLOCK`},
	}
	routeTableRouteRulesRepresentationWithServiceCidrAddingCidrBlock = map[string]interface{}{
		"network_entity_id": Representation{repType: Required, create: `${oci_core_service_gateway.test_service_gateway.id}`},
		"cidr_block":        Representation{repType: Required, create: `${lookup(data.oci_core_services.test_services.services[0], "cidr_block")}`},
		"destination":       Representation{repType: Required, create: `${lookup(data.oci_core_services.test_services.services[0], "cidr_block")}`},
		"destination_type":  Representation{repType: Required, create: `SERVICE_CIDR_BLOCK`},
	}
	routeTableRepresentationWithServiceCidr = getUpdatedRepresentationCopy("route_rules", []RepresentationGroup{
		{Required, routeTableRouteRulesRepresentationWithServiceCidr},
		{Required, routeTableRouteRulesRepresentation}},
		routeTableRepresentation,
	)
	routeTableRepresentationWithServiceCidrAddingCidrBlock = getUpdatedRepresentationCopy("route_rules", []RepresentationGroup{
		{Required, routeTableRouteRulesRepresentationWithServiceCidrAddingCidrBlock},
		{Required, routeTableRouteRulesRepresentation}},
		routeTableRepresentation,
	)

	RouteTableResourceDependencies = VcnResourceConfig + VcnResourceDependencies + `
	resource "oci_core_internet_gateway" "test_network_entity" {
		compartment_id = "${var.compartment_id}"
		vcn_id = "${oci_core_vcn.test_vcn.id}"
		display_name = "-tf-internet-gateway"
	}

	resource "oci_core_drg" "test_drg" {
		#Required
		compartment_id = "${var.compartment_id}"
	}

	resource "oci_core_service_gateway" "test_service_gateway" {
		#Required
		compartment_id = "${var.compartment_id}"
		services {
			service_id = "${lookup(data.oci_core_services.test_services.services[0], "id")}"
		}
		vcn_id = "${oci_core_vcn.test_vcn.id}"
	}
	
	data "oci_core_services" "test_services" {
	}
	`
)

func TestCoreRouteTableResource_basic(t *testing.T) {
	provider := testAccProvider
	config := testProviderConfig()

	compartmentId := getEnvSettingWithBlankDefault("compartment_ocid")
	compartmentIdVariableStr := fmt.Sprintf("variable \"compartment_id\" { default = \"%s\" }\n", compartmentId)

	resourceName := "oci_core_route_table.test_route_table"
	datasourceName := "data.oci_core_route_tables.test_route_tables"

	var resId, resId2 string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: map[string]terraform.ResourceProvider{
			"oci": provider,
		},
		CheckDestroy: testAccCheckCoreRouteTableDestroy,
		Steps: []resource.TestStep{
			// verify create
			{
				Config: config + compartmentIdVariableStr + RouteTableResourceDependencies +
					generateResourceFromRepresentationMap("oci_core_route_table", "test_route_table", Required, Create, routeTableRepresentation),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttr(resourceName, "route_rules.#", "1"),
					CheckResourceSetContainsElementWithProperties(resourceName, "route_rules", map[string]string{
						"cidr_block": "0.0.0.0/0",
					},
						[]string{
							"network_entity_id",
						}),
					resource.TestCheckResourceAttrSet(resourceName, "vcn_id"),

					func(s *terraform.State) (err error) {
						resId, err = fromInstanceState(s, resourceName, "id")
						return err
					},
				),
			},
			// verify update to deprecated cidr_block
			{
				Config: config + compartmentIdVariableStr + RouteTableResourceDependencies +
					generateResourceFromRepresentationMap("oci_core_route_table", "test_route_table", Required, Update, routeTableRepresentation),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_rules.#", "1"),
					CheckResourceSetContainsElementWithProperties(resourceName, "route_rules", map[string]string{"cidr_block": "10.0.0.0/8"}, []string{"network_entity_id"}),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "vcn_id"),

					func(s *terraform.State) (err error) {
						resId2, err = fromInstanceState(s, resourceName, "id")
						if resId != resId2 {
							return fmt.Errorf("Resource recreated when it was supposed to be updated.")
						}
						return err
					},
				),
			},
			// verify update to network_id
			{
				Config: config + compartmentIdVariableStr + RouteTableResourceDependencies +
					generateResourceFromRepresentationMap("oci_core_route_table", "test_route_table", Required, Update,
						getUpdatedRepresentationCopy("route_rules.network_entity_id", Representation{repType: Required, create: `${oci_core_drg.test_drg.id}`},
							routeTableRepresentation,
						)),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_rules.#", "1"),
					CheckResourceSetContainsElementWithProperties(resourceName, "route_rules", map[string]string{"cidr_block": "10.0.0.0/8"}, []string{"network_entity_id"}),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "vcn_id"),

					func(s *terraform.State) (err error) {
						resId2, err = fromInstanceState(s, resourceName, "id")
						if resId != resId2 {
							return fmt.Errorf("Resource recreated when it was supposed to be updated.")
						}
						return err
					},
				),
			},
			// verify create with destination_type
			{
				Config: config + compartmentIdVariableStr + RouteTableResourceDependencies +
					generateResourceFromRepresentationMap("oci_core_route_table", "test_route_table", Required, Create, routeTableRepresentationWithServiceCidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_rules.#", "2"),
					CheckResourceSetContainsElementWithProperties(resourceName, "route_rules", map[string]string{"destination_type": "SERVICE_CIDR_BLOCK"}, []string{"network_entity_id", "destination"}),
					CheckResourceSetContainsElementWithProperties(resourceName, "route_rules", map[string]string{"destination_type": "CIDR_BLOCK", "destination": "0.0.0.0/0"}, []string{"network_entity_id"}),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "vcn_id"),
				),
			},
			// verify update after having a destination_type rule
			{
				Config: config + compartmentIdVariableStr + RouteTableResourceDependencies +
					generateResourceFromRepresentationMap("oci_core_route_table", "test_route_table", Required, Update, routeTableRepresentationWithServiceCidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_rules.#", "2"),
					CheckResourceSetContainsElementWithProperties(resourceName, "route_rules", map[string]string{"destination_type": "SERVICE_CIDR_BLOCK"}, []string{"network_entity_id", "destination"}),
					CheckResourceSetContainsElementWithProperties(resourceName, "route_rules", map[string]string{"destination_type": "CIDR_BLOCK", "destination": "10.0.0.0/8"}, []string{"network_entity_id"}),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "vcn_id"),
				),
			},
			// verify adding cidr_block to a rule that has destination already
			{
				Config: config + compartmentIdVariableStr + RouteTableResourceDependencies +
					generateResourceFromRepresentationMap("oci_core_route_table", "test_route_table", Required, Update, routeTableRepresentationWithServiceCidrAddingCidrBlock),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_rules.#", "2"),
					CheckResourceSetContainsElementWithProperties(resourceName, "route_rules", map[string]string{"destination_type": "SERVICE_CIDR_BLOCK"}, []string{"network_entity_id", "destination"}),
					CheckResourceSetContainsElementWithProperties(resourceName, "route_rules", map[string]string{"destination_type": "CIDR_BLOCK", "destination": "10.0.0.0/8"}, []string{"network_entity_id"}),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "vcn_id"),
				),
			},

			// delete before next create
			{
				Config: config + compartmentIdVariableStr + RouteTableResourceDependencies,
			},
			// verify create with optionals
			{
				Config: config + compartmentIdVariableStr + RouteTableResourceDependencies +
					generateResourceFromRepresentationMap("oci_core_route_table", "test_route_table", Optional, Create, routeTableRepresentation),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttr(resourceName, "defined_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "MyRouteTable"),
					resource.TestCheckResourceAttr(resourceName, "freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_rules.#", "1"),
					CheckResourceSetContainsElementWithProperties(resourceName, "route_rules", map[string]string{
						"cidr_block":       "0.0.0.0/0",
						"destination":      "0.0.0.0/0",
						"destination_type": "CIDR_BLOCK",
					},
						[]string{
							"network_entity_id",
						}),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "vcn_id"),

					func(s *terraform.State) (err error) {
						resId, err = fromInstanceState(s, resourceName, "id")
						return err
					},
				),
			},

			// verify updates to updatable parameters
			{
				Config: config + compartmentIdVariableStr + RouteTableResourceDependencies +
					generateResourceFromRepresentationMap("oci_core_route_table", "test_route_table", Optional, Update, routeTableRepresentation),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttr(resourceName, "defined_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "displayName2"),
					resource.TestCheckResourceAttr(resourceName, "freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_rules.#", "1"),
					CheckResourceSetContainsElementWithProperties(resourceName, "route_rules", map[string]string{
						"cidr_block":       "10.0.0.0/8",
						"destination":      "10.0.0.0/8",
						"destination_type": "CIDR_BLOCK",
					},
						[]string{
							"network_entity_id",
						}),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "vcn_id"),

					func(s *terraform.State) (err error) {
						resId2, err = fromInstanceState(s, resourceName, "id")
						if resId != resId2 {
							return fmt.Errorf("Resource recreated when it was supposed to be updated.")
						}
						return err
					},
				),
			},
			// verify updates to network entity
			{
				Config: config + compartmentIdVariableStr + RouteTableResourceDependencies +
					generateResourceFromRepresentationMap("oci_core_route_table", "test_route_table", Optional, Update,
						getUpdatedRepresentationCopy("route_rules.network_entity_id", Representation{repType: Required, create: `${oci_core_drg.test_drg.id}`},
							routeTableRepresentation,
						)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttr(resourceName, "defined_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "displayName2"),
					resource.TestCheckResourceAttr(resourceName, "freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_rules.#", "1"),
					CheckResourceSetContainsElementWithProperties(resourceName, "route_rules", map[string]string{"cidr_block": "10.0.0.0/8", "destination_type": "CIDR_BLOCK"}, []string{"network_entity_id"}),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttrSet(resourceName, "vcn_id"),

					func(s *terraform.State) (err error) {
						resId2, err = fromInstanceState(s, resourceName, "id")
						if resId != resId2 {
							return fmt.Errorf("Resource recreated when it was supposed to be updated.")
						}
						return err
					},
				),
			},
			// verify datasource
			{
				Config: config +
					generateDataSourceFromRepresentationMap("oci_core_route_tables", "test_route_tables", Optional, Update, routeTableDataSourceRepresentation) +
					compartmentIdVariableStr + RouteTableResourceDependencies +
					generateResourceFromRepresentationMap("oci_core_route_table", "test_route_table", Optional, Update, routeTableRepresentation),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "compartment_id", compartmentId),
					resource.TestCheckResourceAttr(datasourceName, "display_name", "displayName2"),
					resource.TestCheckResourceAttr(datasourceName, "state", "AVAILABLE"),
					resource.TestCheckResourceAttrSet(datasourceName, "vcn_id"),

					resource.TestCheckResourceAttr(datasourceName, "route_tables.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "route_tables.0.compartment_id", compartmentId),
					resource.TestCheckResourceAttr(datasourceName, "route_tables.0.defined_tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "route_tables.0.display_name", "displayName2"),
					resource.TestCheckResourceAttr(datasourceName, "route_tables.0.freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "route_tables.0.id"),
					resource.TestCheckResourceAttr(datasourceName, "route_tables.0.route_rules.#", "1"),
					CheckResourceSetContainsElementWithProperties(datasourceName, "route_tables.0.route_rules", map[string]string{
						"cidr_block":       "10.0.0.0/8",
						"destination":      "10.0.0.0/8",
						"destination_type": "CIDR_BLOCK",
					},
						[]string{
							"network_entity_id",
						}),
					resource.TestCheckResourceAttrSet(datasourceName, "route_tables.0.state"),
					resource.TestCheckResourceAttrSet(datasourceName, "route_tables.0.vcn_id"),
				),
			},
			// verify resource import
			{
				Config:                  config,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
				ResourceName:            resourceName,
			},
		},
	})
}

func testAccCheckCoreRouteTableDestroy(s *terraform.State) error {
	noResourceFound := true
	client := testAccProvider.Meta().(*OracleClients).virtualNetworkClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "oci_core_route_table" {
			noResourceFound = false
			request := oci_core.GetRouteTableRequest{}

			tmp := rs.Primary.ID
			request.RtId = &tmp

			response, err := client.GetRouteTable(context.Background(), request)

			if err == nil {
				deletedLifecycleStates := map[string]bool{
					string(oci_core.RouteTableLifecycleStateTerminated): true,
				}
				if _, ok := deletedLifecycleStates[string(response.LifecycleState)]; !ok {
					//resource lifecycle state is not in expected deleted lifecycle states.
					return fmt.Errorf("resource lifecycle state: %s is not in expected deleted lifecycle states", response.LifecycleState)
				}
				//resource lifecycle state is in expected deleted lifecycle states. continue with next one.
				continue
			}

			//Verify that exception is for '404 not found'.
			if failure, isServiceError := common.IsServiceError(err); !isServiceError || failure.GetHTTPStatusCode() != 404 {
				return err
			}
		}
	}
	if noResourceFound {
		return fmt.Errorf("at least one resource was expected from the state file, but could not be found")
	}

	return nil
}
