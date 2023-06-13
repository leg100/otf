// Code generated by ogen, DO NOT EDIT.

package apigen

import (
	"net/http"
	"net/url"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/conv"
	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/uri"
	"github.com/ogen-go/ogen/validate"
)

// DeleteOrganizationParams is parameters of deleteOrganization operation.
type DeleteOrganizationParams struct {
	// Name of organization to delete.
	Name string
}

func unpackDeleteOrganizationParams(packed middleware.Parameters) (params DeleteOrganizationParams) {
	{
		key := middleware.ParameterKey{
			Name: "name",
			In:   "path",
		}
		params.Name = packed[key].(string)
	}
	return params
}

func decodeDeleteOrganizationParams(args [1]string, argsEscaped bool, r *http.Request) (params DeleteOrganizationParams, _ error) {
	// Decode path: name.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "name",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.Name = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "name",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// DeleteWorkspaceParams is parameters of deleteWorkspace operation.
type DeleteWorkspaceParams struct {
	// Name of workspace to delete.
	ID string
}

func unpackDeleteWorkspaceParams(packed middleware.Parameters) (params DeleteWorkspaceParams) {
	{
		key := middleware.ParameterKey{
			Name: "id",
			In:   "path",
		}
		params.ID = packed[key].(string)
	}
	return params
}

func decodeDeleteWorkspaceParams(args [1]string, argsEscaped bool, r *http.Request) (params DeleteWorkspaceParams, _ error) {
	// Decode path: id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.ID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// ForceUnlockWorkspaceParams is parameters of forceUnlockWorkspace operation.
type ForceUnlockWorkspaceParams struct {
	// ID of workspace to unlock by force.
	ID string
}

func unpackForceUnlockWorkspaceParams(packed middleware.Parameters) (params ForceUnlockWorkspaceParams) {
	{
		key := middleware.ParameterKey{
			Name: "id",
			In:   "path",
		}
		params.ID = packed[key].(string)
	}
	return params
}

func decodeForceUnlockWorkspaceParams(args [1]string, argsEscaped bool, r *http.Request) (params ForceUnlockWorkspaceParams, _ error) {
	// Decode path: id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.ID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// GetOrganizationParams is parameters of getOrganization operation.
type GetOrganizationParams struct {
	// Name of organization to return.
	Name string
}

func unpackGetOrganizationParams(packed middleware.Parameters) (params GetOrganizationParams) {
	{
		key := middleware.ParameterKey{
			Name: "name",
			In:   "path",
		}
		params.Name = packed[key].(string)
	}
	return params
}

func decodeGetOrganizationParams(args [1]string, argsEscaped bool, r *http.Request) (params GetOrganizationParams, _ error) {
	// Decode path: name.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "name",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.Name = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "name",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// GetWorkspaceParams is parameters of getWorkspace operation.
type GetWorkspaceParams struct {
	// ID of workspace to return.
	ID string
}

func unpackGetWorkspaceParams(packed middleware.Parameters) (params GetWorkspaceParams) {
	{
		key := middleware.ParameterKey{
			Name: "id",
			In:   "path",
		}
		params.ID = packed[key].(string)
	}
	return params
}

func decodeGetWorkspaceParams(args [1]string, argsEscaped bool, r *http.Request) (params GetWorkspaceParams, _ error) {
	// Decode path: id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.ID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// GetWorkspaceByNameParams is parameters of getWorkspaceByName operation.
type GetWorkspaceByNameParams struct {
	// Name of workspace's organization.
	Organization string
	// Name of workspace to return.
	Name string
}

func unpackGetWorkspaceByNameParams(packed middleware.Parameters) (params GetWorkspaceByNameParams) {
	{
		key := middleware.ParameterKey{
			Name: "organization",
			In:   "path",
		}
		params.Organization = packed[key].(string)
	}
	{
		key := middleware.ParameterKey{
			Name: "name",
			In:   "path",
		}
		params.Name = packed[key].(string)
	}
	return params
}

func decodeGetWorkspaceByNameParams(args [2]string, argsEscaped bool, r *http.Request) (params GetWorkspaceByNameParams, _ error) {
	// Decode path: organization.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "organization",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.Organization = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "organization",
			In:   "path",
			Err:  err,
		}
	}
	// Decode path: name.
	if err := func() error {
		param := args[1]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[1])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "name",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.Name = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "name",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// ListWorkspacesParams is parameters of listWorkspaces operation.
type ListWorkspacesParams struct {
	// The organization name of the workspaces to list.
	Organization string
	// The page number to request.
	PageNumber OptInt
	// The number of items to be returned per page.
	PageSize OptInt
	// Search workspace by name.
	SearchName OptString
	// Search workspace by tags.
	SearchTags []string
}

func unpackListWorkspacesParams(packed middleware.Parameters) (params ListWorkspacesParams) {
	{
		key := middleware.ParameterKey{
			Name: "organization",
			In:   "path",
		}
		params.Organization = packed[key].(string)
	}
	{
		key := middleware.ParameterKey{
			Name: "page[number]",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.PageNumber = v.(OptInt)
		}
	}
	{
		key := middleware.ParameterKey{
			Name: "page[size]",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.PageSize = v.(OptInt)
		}
	}
	{
		key := middleware.ParameterKey{
			Name: "search[name]",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.SearchName = v.(OptString)
		}
	}
	{
		key := middleware.ParameterKey{
			Name: "search[tags]",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.SearchTags = v.([]string)
		}
	}
	return params
}

func decodeListWorkspacesParams(args [1]string, argsEscaped bool, r *http.Request) (params ListWorkspacesParams, _ error) {
	q := uri.NewQueryDecoder(r.URL.Query())
	// Decode path: organization.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "organization",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.Organization = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "organization",
			In:   "path",
			Err:  err,
		}
	}
	// Decode query: page[number].
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "page[number]",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotPageNumberVal int
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToInt(val)
					if err != nil {
						return err
					}

					paramsDotPageNumberVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.PageNumber.SetTo(paramsDotPageNumberVal)
				return nil
			}); err != nil {
				return err
			}
			if err := func() error {
				if params.PageNumber.Set {
					if err := func() error {
						if err := (validate.Int{
							MinSet:        true,
							Min:           1,
							MaxSet:        false,
							Max:           0,
							MinExclusive:  false,
							MaxExclusive:  false,
							MultipleOfSet: false,
							MultipleOf:    0,
						}).Validate(int64(params.PageNumber.Value)); err != nil {
							return errors.Wrap(err, "int")
						}
						return nil
					}(); err != nil {
						return err
					}
				}
				return nil
			}(); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "page[number]",
			In:   "query",
			Err:  err,
		}
	}
	// Set default value for query: page[size].
	{
		val := int(100)
		params.PageSize.SetTo(val)
	}
	// Decode query: page[size].
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "page[size]",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotPageSizeVal int
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToInt(val)
					if err != nil {
						return err
					}

					paramsDotPageSizeVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.PageSize.SetTo(paramsDotPageSizeVal)
				return nil
			}); err != nil {
				return err
			}
			if err := func() error {
				if params.PageSize.Set {
					if err := func() error {
						if err := (validate.Int{
							MinSet:        true,
							Min:           1,
							MaxSet:        true,
							Max:           100,
							MinExclusive:  false,
							MaxExclusive:  false,
							MultipleOfSet: false,
							MultipleOf:    0,
						}).Validate(int64(params.PageSize.Value)); err != nil {
							return errors.Wrap(err, "int")
						}
						return nil
					}(); err != nil {
						return err
					}
				}
				return nil
			}(); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "page[size]",
			In:   "query",
			Err:  err,
		}
	}
	// Decode query: search[name].
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "search[name]",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotSearchNameVal string
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToString(val)
					if err != nil {
						return err
					}

					paramsDotSearchNameVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.SearchName.SetTo(paramsDotSearchNameVal)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "search[name]",
			In:   "query",
			Err:  err,
		}
	}
	// Decode query: search[tags].
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "search[tags]",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				return d.DecodeArray(func(d uri.Decoder) error {
					var paramsDotSearchTagsVal string
					if err := func() error {
						val, err := d.DecodeValue()
						if err != nil {
							return err
						}

						c, err := conv.ToString(val)
						if err != nil {
							return err
						}

						paramsDotSearchTagsVal = c
						return nil
					}(); err != nil {
						return err
					}
					params.SearchTags = append(params.SearchTags, paramsDotSearchTagsVal)
					return nil
				})
			}); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "search[tags]",
			In:   "query",
			Err:  err,
		}
	}
	return params, nil
}

// LockWorkspaceParams is parameters of lockWorkspace operation.
type LockWorkspaceParams struct {
	// ID of workspace to lock.
	ID string
}

func unpackLockWorkspaceParams(packed middleware.Parameters) (params LockWorkspaceParams) {
	{
		key := middleware.ParameterKey{
			Name: "id",
			In:   "path",
		}
		params.ID = packed[key].(string)
	}
	return params
}

func decodeLockWorkspaceParams(args [1]string, argsEscaped bool, r *http.Request) (params LockWorkspaceParams, _ error) {
	// Decode path: id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.ID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// UnlockWorkspaceParams is parameters of unlockWorkspace operation.
type UnlockWorkspaceParams struct {
	// ID of workspace to unlock.
	ID string
}

func unpackUnlockWorkspaceParams(packed middleware.Parameters) (params UnlockWorkspaceParams) {
	{
		key := middleware.ParameterKey{
			Name: "id",
			In:   "path",
		}
		params.ID = packed[key].(string)
	}
	return params
}

func decodeUnlockWorkspaceParams(args [1]string, argsEscaped bool, r *http.Request) (params UnlockWorkspaceParams, _ error) {
	// Decode path: id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.ID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// UpdateWorkspaceParams is parameters of updateWorkspace operation.
type UpdateWorkspaceParams struct {
	// ID of workspace to update.
	ID string
}

func unpackUpdateWorkspaceParams(packed middleware.Parameters) (params UpdateWorkspaceParams) {
	{
		key := middleware.ParameterKey{
			Name: "id",
			In:   "path",
		}
		params.ID = packed[key].(string)
	}
	return params
}

func decodeUpdateWorkspaceParams(args [1]string, argsEscaped bool, r *http.Request) (params UpdateWorkspaceParams, _ error) {
	// Decode path: id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.ID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// UpdateWorkspaceByNameParams is parameters of updateWorkspaceByName operation.
type UpdateWorkspaceByNameParams struct {
	// Name of workspace's organization.
	Organization string
	// Name of workspace to update.
	Name string
}

func unpackUpdateWorkspaceByNameParams(packed middleware.Parameters) (params UpdateWorkspaceByNameParams) {
	{
		key := middleware.ParameterKey{
			Name: "organization",
			In:   "path",
		}
		params.Organization = packed[key].(string)
	}
	{
		key := middleware.ParameterKey{
			Name: "name",
			In:   "path",
		}
		params.Name = packed[key].(string)
	}
	return params
}

func decodeUpdateWorkspaceByNameParams(args [2]string, argsEscaped bool, r *http.Request) (params UpdateWorkspaceByNameParams, _ error) {
	// Decode path: organization.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "organization",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.Organization = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "organization",
			In:   "path",
			Err:  err,
		}
	}
	// Decode path: name.
	if err := func() error {
		param := args[1]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[1])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "name",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.Name = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "name",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}
