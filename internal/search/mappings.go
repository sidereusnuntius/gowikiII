package search

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/v2/mapping"
)

func articleIndexMapping() mapping.IndexMapping {
	articleMapping := bleve.NewDocumentMapping()

	textfieldMapping := bleve.NewTextFieldMapping()
	textfieldMapping.Analyzer = simple.Name

	articleMapping.AddFieldMappingsAt("title", textfieldMapping)
	articleMapping.AddFieldMappingsAt("content", textfieldMapping)
	articleMapping.AddFieldMappingsAt("summary", textfieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("article", articleMapping)
	return indexMapping
}
