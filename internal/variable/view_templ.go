// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.833
package variable

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"github.com/leg100/otf/internal/http/html/components"
	"github.com/leg100/otf/internal/http/html/components/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/workspace"
)

func new(ws *workspace.Workspace) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var2 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
			if !templ_7745c5c3_IsBuffer {
				defer func() {
					templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err == nil {
						templ_7745c5c3_Err = templ_7745c5c3_BufErr
					}
				}()
			}
			ctx = templ.InitializeContext(ctx)
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<span class=\"text-xl\">Add a new workspace variable.</span>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = variableForm(variableFormProps{
				ws:         ws,
				formAction: paths.CreateVariable(ws.ID.String()),
			}).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return nil
		})
		templ_7745c5c3_Err = components.Layout(components.LayoutProps{
			Title: "new variable",
			Breadcrumbs: []components.Breadcrumb{
				{Name: "workspaces", Link: paths.Workspaces(ws.Organization)},
				{Name: ws.Name, Link: paths.Workspace(ws.ID.String())},
				{Name: "new"},
			},
		}).Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

type listWorkspaceVariablesProps struct {
	workspace.WorkspacePage
	WorkspaceVariableTable workspaceVariableTable
	VariableSetTables      []setVariableTable
	CanCreateVariable      bool
	CanDeleteVariable      bool
	CanUpdateWorkspace     bool
}

func listWorkspaceVariables(props listWorkspaceVariablesProps) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var3 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var3 == nil {
			templ_7745c5c3_Var3 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		return nil
	})
}

func editWorkspaceVariable(props variableFormProps) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var4 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var4 == nil {
			templ_7745c5c3_Var4 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		return nil
	})
}

type listVariableSetsProps struct {
	WorkspaceVariableTable workspaceVariableTable
	VariableSetTables      []setVariableTable
	CanCreateVariable      bool
	CanDeleteVariable      bool
	CanUpdateWorkspace     bool
}

func listVariableSets(props listWorkspaceVariablesProps) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var5 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var5 == nil {
			templ_7745c5c3_Var5 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		return nil
	})
}

type newVariableSetProps struct {
	WorkspaceVariableTable workspaceVariableTable
	VariableSetTables      []setVariableTable
	CanCreateVariable      bool
	CanDeleteVariable      bool
	CanUpdateWorkspace     bool
}

func newVariableSet(props newVariableSetProps) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var6 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var6 == nil {
			templ_7745c5c3_Var6 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		return nil
	})
}

type editVariableSetProps struct {
	organization.OrganizationPage
	*VariableSet
	EditMode            bool
	FormAction          string
	AvailableWorkspaces []workspaceInfo
	ExistingWorkspaces  []workspaceInfo
	CanCreateVariable   bool
	CanDeleteVariable   bool
	VariableTable       setVariableTable
}

func editVariableSet(props editVariableSetProps) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var7 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var7 == nil {
			templ_7745c5c3_Var7 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		return nil
	})
}

type newVariableSetVariableProps struct {
	organization.OrganizationPage
	*VariableSet
	EditMode            bool
	FormAction          string
	AvailableWorkspaces []workspaceInfo
	ExistingWorkspaces  []workspaceInfo
	CanCreateVariable   bool
	CanDeleteVariable   bool
	VariableTable       setVariableTable
}

func newVariableSetVariable(props newVariableSetVariableProps) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var8 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var8 == nil {
			templ_7745c5c3_Var8 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		return nil
	})
}

func editVariableSetVariable(props newVariableSetVariableProps) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var9 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var9 == nil {
			templ_7745c5c3_Var9 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		return nil
	})
}

type variableFormProps struct {
	variable   *Variable
	ws         *workspace.Workspace
	editMode   bool
	formAction templ.SafeURL
}

func variableForm(props variableFormProps) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var10 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var10 == nil {
			templ_7745c5c3_Var10 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "<form class=\"flex flex-col gap-5\" action=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var11 templ.SafeURL = props.formAction
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string(templ_7745c5c3_Var11)))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "\" method=\"POST\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if props.variable != nil {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "<div class=\"field\"><label class=\"font-semibold\" for=\"key\">Key</label> <input class=\"text-input\" type=\"text\" name=\"key\" id=\"key\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var12 string
			templ_7745c5c3_Var12, templ_7745c5c3_Err = templ.JoinStringErrs(props.variable.Key)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/variable/view.templ`, Line: 109, Col: 88}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var12))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "\" required placeholder=\"key\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if props.variable.Sensitive && props.editMode {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, " disabled")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 7, "></div><div class=\"field\"><label class=\"font-semibold\" for=\"value\">Value</label> <textarea class=\"text-input\" type=\"text\" name=\"value\" id=\"value\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if props.variable.Sensitive {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 8, " placeholder=\"hidden\"")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 9, ">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if props.variable.Sensitive {
				var templ_7745c5c3_Var13 string
				templ_7745c5c3_Var13, templ_7745c5c3_Err = templ.JoinStringErrs(props.variable.Value)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/variable/view.templ`, Line: 123, Col: 28}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var13))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 10, "</textarea></div><fieldset class=\"border border-slate-900 px-3 py-3 flex flex-col gap-2\"><legend>Category</legend><div class=\"form-checkbox\"><input type=\"radio\" name=\"category\" id=\"terraform\" value=\"terraform\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if props.variable.Category == "terraform" {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 11, " checked")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if props.variable.Sensitive && props.editMode {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 12, " disabled")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 13, " required> <label for=\"terraform\">Terraform</label> <span class=\"description\">These variables should match the declarations in your configuration. Click the HCL box to use interpolation or set a non-string value.</span></div><div class=\"form-checkbox\"><input type=\"radio\" name=\"category\" id=\"env\" value=\"env\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if props.variable.Category == "env" {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 14, " checked")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if props.variable.Sensitive && props.editMode {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 15, " disabled")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 16, " required> <label for=\"env\">Environment variable</label> <span class=\"description\">These variables are available in the Terraform runtime environment.</span></div></fieldset><div class=\"form-checkbox\"><input class=\"\" type=\"checkbox\" name=\"hcl\" id=\"hcl\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if props.variable.HCL {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 17, " checked")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if props.variable.Sensitive && props.editMode {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 18, " disabled")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 19, "> <label for=\"hcl\">HCL</label> <span class=\"description\">Parse this field as HashiCorp Configuration Language (HCL). This allows you to interpolate values at runtime.</span></div><div class=\"form-checkbox\"><input type=\"checkbox\" name=\"sensitive\" id=\"sensitive\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if props.variable.Sensitive {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 20, " checked")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if props.variable.Sensitive && props.editMode {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 21, " disabled")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 22, "> <label for=\"sensitive\">Sensitive</label> <span class=\"description\">Sensitive variables are never shown in the UI or API. They may appear in Terraform logs if your configuration is designed to output them.</span></div><div class=\"field\"><label class=\"font-semibold\" for=\"description\">Description</label> <input class=\"text-input\" type=\"text\" class=\"freeform\" name=\"description\" id=\"description\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var14 string
			templ_7745c5c3_Var14, templ_7745c5c3_Err = templ.JoinStringErrs(props.variable.Description)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/variable/view.templ`, Line: 152, Col: 129}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var14))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 23, "\" placeholder=\"description (optional)\"></div><div><button class=\"btn\" id=\"save-variable-button\">Save variable</button></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 24, "</form><hr class=\"my-5\"><div class=\"flex flex-col gap-2\"><h4 class=\"font-semibold\">Sensitive variables</h4>Sensitive variables are never shown in the UI or API. Other than changing its value, you cannot make changes to a sensitive variable; you'll need to delete and re-create the variable.</div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
