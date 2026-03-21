package parser

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
)

type Plan struct {
	Nodes map[string]*Node
	Roots []string // Node IDs that have no dependencies and can be built immediately
}

type Node struct {
	ID        string // Using img.UniqName()
	Image     *image.Image
	DependsOn []string // IDs of nodes this node depends on
}

// GeneratePlan builds a Directed Acyclic Graph of images based on config and parsed FROM statements.
func GeneratePlan(cfg *config.Config, flags *config.Flags) (*Plan, error) {
	plan := &Plan{
		Nodes: make(map[string]*Node),
		Roots: make([]string, 0),
	}

	var chronologicalNodes []*Node

	// 1. Generate all *image.Image instances
	for _, name := range cfg.ImageOrder {
		// Build only what's provided by --image flag (single image)
		if flags.Image != "" && name != flags.Image {
			continue
		}

		imageCfg := cfg.Images[name]
		log.Debug().Str("image", name).Interface("config", imageCfg).Msg("Parsing")
		log.Debug().Interface("excludes", imageCfg.Excludes).Msg("Excluded config sets")

		combinations := GenerateVariableCombinations(imageCfg.Variables)
		for _, rawConfigSet := range combinations {
			img := image.From(name, cfg, rawConfigSet, flags)

			// skip excluded config sets
			if isExcluded(img.ConfigSet(), imageCfg.Excludes) {
				log.Warn().Interface("config set", img.Representation()).Interface("excludes", imageCfg.Excludes).Msg("Skipping excluded")
				continue
			}

			if err := img.Validate(); err != nil {
				return nil, err
			}

			if err := img.Render(); err != nil {
				return nil, err
			}
			log.Debug().Interface("config set", img.Representation()).Msg("Generated")

			id := img.UniqName()
			node := &Node{
				ID:        id,
				Image:     img,
				DependsOn: []string{},
			}
			plan.Nodes[id] = node
			chronologicalNodes = append(chronologicalNodes, node)
		}
	}

	// 1.5 Deduplicate alias tags (Last Write Wins) based on chronological order
	finalTagOwners := make(map[string]string) // tag -> NodeID
	// Forward pass to record the *last* owner of each tag
	for _, node := range chronologicalNodes {
		for _, tag := range node.Image.Tags() {
			finalTagOwners[tag] = node.ID
		}
	}

	// Prune overridden tags
	for _, node := range chronologicalNodes {
		var keptOriginalTags []string
		originalTags := node.Image.OriginalTags()
		fullyQualifiedTags := node.Image.Tags()
		for i, tag := range fullyQualifiedTags {
			if finalTagOwners[tag] == node.ID {
				keptOriginalTags = append(keptOriginalTags, originalTags[i])
			} else {
				log.Debug().Str("tag", tag).Str("image", node.ID).Msg("Tag deduplicated (overwritten by later matrix configuration)")
			}
		}
		node.Image.SetOriginalTags(keptOriginalTags)
	}

	// 2. Identify dependencies between generated images
	for _, node := range plan.Nodes {
		deps, err := node.Image.ExtractFromDependencies()
		if err != nil {
			log.Warn().Err(err).Str("image", node.ID).Msg("Failed to extract FROM dependencies. Assuming no inner dependencies.")
			deps = []string{}
		}

		for _, depRef := range deps {
			// Find which node provides this tag
			for providerID, providerNode := range plan.Nodes {
				if providerID == node.ID {
					continue
				}

				// Does providerNode generate the exact tag?
				provides := false
				for _, tag := range providerNode.Image.Tags() {
					if tag == depRef {
						provides = true
						break
					}
				}

				if provides {
					node.DependsOn = append(node.DependsOn, providerID)
					// Note: Since tags can be aliases, we just need one edge per provider image.
					break
				}
			}
		}
	}

	// 3. Validate DAG and find Roots
	if err := plan.validateNoCycles(); err != nil {
		return nil, err
	}

	// Find Roots (nodes with DependsOn == 0)
	for id, node := range plan.Nodes {
		if len(node.DependsOn) == 0 {
			plan.Roots = append(plan.Roots, id)
		}
	}

	return plan, nil
}

// validateNoCycles checks for cyclic dependencies using a depth-first search.
func (p *Plan) validateNoCycles() error {
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(id string) error
	visit = func(id string) error {
		if visiting[id] {
			return fmt.Errorf("cyclic dependency detected involving image %s", id)
		}
		if visited[id] {
			return nil
		}
		visiting[id] = true

		node := p.Nodes[id]
		if node != nil {
			for _, depID := range node.DependsOn {
				if err := visit(depID); err != nil {
					return err
				}
			}
		}

		visiting[id] = false
		visited[id] = true
		return nil
	}

	for id := range p.Nodes {
		if err := visit(id); err != nil {
			return err
		}
	}

	return nil
}

// Layers performs a topological sort and returns a slice of layers, where each layer
// is a slice of Nodes that can be built in parallel.
func (p *Plan) Layers() [][]*Node {
	var layers [][]*Node

	// Track nodes we've built
	built := make(map[string]bool)

	for len(built) < len(p.Nodes) {
		var currentLayer []*Node

		for id, node := range p.Nodes {
			if built[id] {
				continue
			}

			// Check if all dependencies are built
			canBuild := true
			for _, depID := range node.DependsOn {
				if !built[depID] {
					canBuild = false
					break
				}
			}

			if canBuild {
				currentLayer = append(currentLayer, node)
			}
		}

		// If we found nodes, add them to the layer and mark as built
		if len(currentLayer) > 0 {
			layers = append(layers, currentLayer)
			for _, node := range currentLayer {
				built[node.ID] = true
			}
		} else {
			// This should not happen if validateNoCycles passed, but just in case
			break
		}
	}

	return layers
}
