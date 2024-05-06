// In Kubernetes, a "chart" typically refers to a package of pre-configured Kubernetes resources that can be easily deployed onto a Kubernetes cluster. 
// Charts are used specifically within the context of Helm, which is a package manager for Kubernetes. Helm charts encapsulate all the necessary Kubernetes
// manifests, configurations, and dependencies needed to deploy a particular application or service.

// Here are some key points about Helm charts and their usage in Kubernetes:

// 1) Packaged Resources: A Helm chart is a collection of files organized into a specific directory structure. These files include Kubernetes manifest files (YAML), 
//    templates, configuration files, and any other resources required by the application.
// 2) Templating: Helm supports templating using the Go template language. Templates allow users to parameterize their Kubernetes manifests, making
//    them more flexible and reusable. Users can define values that can be dynamically substituted into the manifest files during deployment.
// 3) Dependency Management: Helm charts can have dependencies on other charts. Helm automatically manages these dependencies, making it easier to 
//    package and distribute complex applications composed of multiple components.
// 4) Versioning and Release Management: Helm provides versioning and release management features, allowing users to manage different versions 
//    of their applications and perform upgrades and rollbacks easily.
// 5) Repository: Helm charts can be published to and retrieved from chart repositories. Public repositories like the official Helm Hub or private 
//    repositories can be used to share and distribute charts across teams and organizations.

// Overall, Helm charts simplify the process of deploying and managing applications on Kubernetes by providing a standardized way to package, version, 
// and distribute Kubernetes resources and configurations. They promote best practices for application deployment and infrastructure management in Kubernetes environments.

// NOTE: not really sure if Helm charts are relevant to the simulator. To be checked.

package chart

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	simontype "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/type"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/releaseutil"
)

// ProcessChart parses chart to /tmp/charts
func ProcessChart(name string, chartPath string) ([]string, error) {
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}
	chartRequested.Metadata.Name = name

	if err := checkIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	// TODO
	var vals map[string]interface{}
	if err := chartutil.ProcessDependencies(chartRequested, vals); err != nil {
		return nil, err
	}

	valuesToRender, err := ToRenderValues(chartRequested, vals)
	if err != nil {
		return nil, err
	}

	return renderResources(chartRequested, valuesToRender, true)
}

// checkIfInstallable validates if a chart can be installed
// Application chart type is only installable
func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return fmt.Errorf("%s charts are not installable", ch.Metadata.Type)
}

// ToRenderValues composes the struct from the data coming from the Releases, Charts and Values files
func ToRenderValues(chrt *chart.Chart, chrtVals map[string]interface{}) (chartutil.Values, error) {

	top := map[string]interface{}{
		"Chart": chrt.Metadata,
		"Release": map[string]interface{}{
			"Name":      chrt.Name(),
			"Namespace": "default",
			"Revision":  1,
			"Service":   "Helm",
		},
	}

	vals, err := chartutil.CoalesceValues(chrt, chrtVals)
	if err != nil {
		return top, err
	}

	if err := chartutil.ValidateAgainstSchema(chrt, vals); err != nil {
		errFmt := "values don't meet the specifications of the schema(s) in the following chart(s):\n%s"
		return top, fmt.Errorf(errFmt, err.Error())
	}

	top["Values"] = vals
	return top, nil
}

func renderResources(ch *chart.Chart, values chartutil.Values, subNotes bool) ([]string, error) {
	files, err := engine.Render(ch, values)
	if err != nil {
		return nil, err
	}

	// NOTES.txt gets rendered like all the other files, but because it's not a hook nor a resource,
	// pull it out of here into a separate file so that we can actually use the output of the rendered
	// text file. We have to spin through this map because the file contains path information, so we
	// look for terminating NOTES.txt. We also remove it from the files so that we don't have to skip
	// it in the sortHooks.
	var notesBuffer bytes.Buffer
	for k, v := range files {
		if strings.HasSuffix(k, simontype.NotesFileSuffix) {
			if subNotes || (k == path.Join(ch.Name(), "templates", simontype.NotesFileSuffix)) {
				// If buffer contains data, add newline before adding more
				if notesBuffer.Len() > 0 {
					notesBuffer.WriteString("\n")
				}
				notesBuffer.WriteString(v)
			}
			delete(files, k)
		}
	}

	// Sort hooks, manifests, and partials. Only hooks and manifests are returned,
	// as partials are not used after renderer.Render. Empty manifests are also
	// removed here.
	var yamlStr []string
	_, manifests, err := releaseutil.SortManifests(files, []string{}, releaseutil.InstallOrder)
	if err != nil {
		return nil, err
	}
	for _, item := range manifests {
		yamlStr = append(yamlStr, item.Content)
	}

	return yamlStr, nil
}
