package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type HistoryCollector struct {
	*object.HistoryCollector
}

func main() {
	// Creating a connection context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parsing URL
	url, err := url.Parse("https://administrator%40vsphere.local:<password>@<yourvcenter>/sdk")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	// Connecting to vCenter
	c, err := govmomi.NewClient(ctx, url, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	// Get Datacenters
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datacenter"}, true)
	if err != nil {
		log.Fatal(err)
	}

	var dcs []mo.Datacenter
	err = v.Retrieve(ctx, []string{"Datacenter"}, nil, &dcs)
	if err != nil {
		log.Fatal(err)
	}
	for _, dc := range dcs {
		if dc.Name == "Datacenter" {
			dcRef := dc.Reference()
			taskManager := c.ServiceContent.TaskManager
			filter := types.TaskFilterSpec{
				Entity: &types.TaskFilterSpecByEntity{
					Entity:    dcRef,
					Recursion: types.TaskFilterSpecRecursionOptionAll,
				},
			}
			req := types.CreateCollectorForTasks{
				This:   taskManager.Reference(),
				Filter: filter,
			}
			res, err := methods.CreateCollectorForTasks(ctx, m.Client(), &req)
			if err != nil {
				log.Fatal(err)
			}
			h := &HistoryCollector{
				HistoryCollector: object.NewHistoryCollector(c.Client, res.Returnval),
			}
			var taskh mo.TaskHistoryCollector
			h.Properties(ctx, h.Reference(), []string{"latestPage"}, &taskh)
			for _, task := range taskh.LatestPage {
				fmt.Printf("%#v\n", task)
			}
		}
	}

	defer v.Destroy(ctx)

}
