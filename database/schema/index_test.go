package schema

import (
	"strings"
	"testing"

	"github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/neat/database/schema/constants"
)

// --- IndexDefinition tests ---

func TestIndexDefinitionAlgorithm(t *testing.T) {
	cmd := &schema.Command{}
	def := NewIndexDefinition(cmd)

	result := def.Algorithm("BTREE")

	if result == nil {
		t.Fatal("Expected non-nil IndexDefinition from Algorithm")
	}
	if cmd.Algorithm != "BTREE" {
		t.Errorf("Expected Algorithm 'BTREE', got %q", cmd.Algorithm)
	}
}

func TestIndexDefinitionName(t *testing.T) {
	cmd := &schema.Command{}
	def := NewIndexDefinition(cmd)

	result := def.Name("idx_users_email")

	if result == nil {
		t.Fatal("Expected non-nil IndexDefinition from Name")
	}
	if cmd.Index != "idx_users_email" {
		t.Errorf("Expected Index name 'idx_users_email', got %q", cmd.Index)
	}
}

func TestIndexDefinitionDeferrable(t *testing.T) {
	cmd := &schema.Command{}
	def := NewIndexDefinition(cmd)

	result := def.Deferrable()

	if result == nil {
		t.Fatal("Expected non-nil IndexDefinition from Deferrable")
	}
	if cmd.Deferrable == nil || !*cmd.Deferrable {
		t.Error("Expected Deferrable to be set to true")
	}
}

func TestIndexDefinitionInitiallyImmediate(t *testing.T) {
	cmd := &schema.Command{}
	def := NewIndexDefinition(cmd)

	result := def.InitiallyImmediate()

	if result == nil {
		t.Fatal("Expected non-nil IndexDefinition from InitiallyImmediate")
	}
	if cmd.InitiallyImmediate == nil || !*cmd.InitiallyImmediate {
		t.Error("Expected InitiallyImmediate to be set to true")
	}
}

func TestIndexDefinitionLanguage(t *testing.T) {
	cmd := &schema.Command{}
	def := NewIndexDefinition(cmd)

	result := def.Language("english")

	if result == nil {
		t.Fatal("Expected non-nil IndexDefinition from Language")
	}
	if cmd.Language != "english" {
		t.Errorf("Expected Language 'english', got %q", cmd.Language)
	}
}

// --- Blueprint index creation tests ---

func TestBlueprintIndexCreation(t *testing.T) {
	bp := NewBlueprint(nil, "", "users")

	def := bp.Index("email")

	if def == nil {
		t.Fatal("Expected non-nil IndexDefinition from Index()")
	}
	if !bp.HasCommand(constants.CommandIndex) {
		t.Error("Expected blueprint to have an index command")
	}
}

func TestBlueprintUniqueIndexCreation(t *testing.T) {
	bp := NewBlueprint(nil, "", "users")

	def := bp.Unique("email")

	if def == nil {
		t.Fatal("Expected non-nil IndexDefinition from Unique()")
	}
	if !bp.HasCommand(constants.CommandUnique) {
		t.Error("Expected blueprint to have a unique command")
	}
}

func TestBlueprintCompositeIndexCreation(t *testing.T) {
	bp := NewBlueprint(nil, "", "orders")

	def := bp.Index("user_id", "status")

	if def == nil {
		t.Fatal("Expected non-nil IndexDefinition from composite Index()")
	}

	cmds := bp.GetCommands()
	var found *schema.Command
	for _, c := range cmds {
		if c.Name == constants.CommandIndex {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatal("Expected to find index command in blueprint")
	}
	if len(found.Columns) != 2 {
		t.Errorf("Expected 2 columns in composite index, got %d", len(found.Columns))
	}
}

func TestBlueprintIndexNaming(t *testing.T) {
	bp := NewBlueprint(nil, "", "users")
	bp.Index("email")

	cmds := bp.GetCommands()
	var found *schema.Command
	for _, c := range cmds {
		if c.Name == constants.CommandIndex {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatal("Expected to find index command")
	}

	// Default name should contain table name, column, and type
	if !strings.Contains(found.Index, "users") {
		t.Errorf("Expected index name to contain table 'users', got %q", found.Index)
	}
	if !strings.Contains(found.Index, "email") {
		t.Errorf("Expected index name to contain column 'email', got %q", found.Index)
	}
	if !strings.Contains(found.Index, "index") {
		t.Errorf("Expected index name to contain type 'index', got %q", found.Index)
	}
}

func TestBlueprintIndexCustomName(t *testing.T) {
	bp := NewBlueprint(nil, "", "users")
	bp.Index("email").Name("custom_idx")

	cmds := bp.GetCommands()
	var found *schema.Command
	for _, c := range cmds {
		if c.Name == constants.CommandIndex {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatal("Expected to find index command")
	}
	if found.Index != "custom_idx" {
		t.Errorf("Expected custom index name 'custom_idx', got %q", found.Index)
	}
}

func TestBlueprintDropIndex(t *testing.T) {
	bp := NewBlueprint(nil, "", "users")
	bp.DropIndex("email")

	if !bp.HasCommand(constants.CommandDropIndex) {
		t.Error("Expected blueprint to have a dropIndex command")
	}
}

func TestBlueprintDropIndexByName(t *testing.T) {
	bp := NewBlueprint(nil, "", "users")
	bp.DropIndexByName("idx_users_email")

	cmds := bp.GetCommands()
	var found *schema.Command
	for _, c := range cmds {
		if c.Name == constants.CommandDropIndex {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatal("Expected to find dropIndex command")
	}
	if found.Index != "idx_users_email" {
		t.Errorf("Expected drop index name 'idx_users_email', got %q", found.Index)
	}
}

func TestBlueprintDropUniqueIndex(t *testing.T) {
	bp := NewBlueprint(nil, "", "users")
	bp.DropUnique("email")

	if !bp.HasCommand(constants.CommandDropUnique) {
		t.Error("Expected blueprint to have a dropUnique command")
	}
}

func TestBlueprintIndexListing(t *testing.T) {
	bp := NewBlueprint(nil, "", "products")
	bp.Index("name")
	bp.Unique("sku")
	bp.Index("category_id", "status")

	cmds := bp.GetCommands()

	var indexCount int
	for _, c := range cmds {
		if c.Name == constants.CommandIndex || c.Name == constants.CommandUnique {
			indexCount++
		}
	}

	if indexCount != 3 {
		t.Errorf("Expected 3 index-related commands, got %d", indexCount)
	}
}

func TestBlueprintIndexNameWithPrefix(t *testing.T) {
	bp := NewBlueprint(nil, "app_", "orders")
	bp.Index("user_id")

	cmds := bp.GetCommands()
	var found *schema.Command
	for _, c := range cmds {
		if c.Name == constants.CommandIndex {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatal("Expected to find index command")
	}
	if !strings.Contains(found.Index, "app_") {
		t.Errorf("Expected index name to contain prefix 'app_', got %q", found.Index)
	}
}
