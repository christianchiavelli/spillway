package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-faster/jx"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/validate"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/httpapi/oas"
	"github.com/christianchiavelli/spillway/services/api/internal/app/identity"
	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/alert"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

// NewError turns the error of a handler or of the security handler into the
// problem the API answers with. Causes are told apart with errors.Is and
// errors.As, never by their text, so wrapping an error on its way up keeps
// its meaning.
func (h *handler) NewError(ctx context.Context, err error) *oas.ProblemStatusCodeWithHeaders {
	var (
		invalid   *domain.ValidationError
		conflict  *domain.ConflictError
		throttled *identity.ThrottledError
	)
	switch {
	case errors.As(err, &invalid):
		p := problem(http.StatusUnprocessableEntity, oas.ProblemCodeValidationFailed, "Some fields break the rules.")
		for _, v := range invalid.Violations {
			p.Response.Errors = append(p.Response.Errors, oas.FieldError{Pointer: "/" + v.Field, Code: string(v.Code)})
		}
		return p
	case errors.Is(err, alert.ErrLimitReached):
		return problem(http.StatusConflict, oas.ProblemCodeAlertLimitReached, "The account already holds the most alerts it can.")
	case errors.As(err, &conflict):
		p := problem(http.StatusConflict, oas.ProblemCodeConflict, "Another record already uses that value.")
		p.Response.Errors = []oas.FieldError{{Pointer: "/" + conflict.Field, Code: string(conflict.Code)}}
		return p
	case errors.As(err, &throttled):
		p := problem(http.StatusTooManyRequests, oas.ProblemCodeRateLimited, "Too many attempts. Wait before trying again.")
		p.RetryAfter = oas.NewOptInt(int(math.Ceil(throttled.RetryAfter.Seconds())))
		return p
	case errors.Is(err, identity.ErrInvalidCredentials):
		// 400, as an OAuth token endpoint answers invalid_grant: a client that
		// treats 401 as "sign in again" must not loop on the sign-in itself.
		return problem(http.StatusBadRequest, oas.ProblemCodeInvalidCredentials, "The email or the password is wrong.")
	case errors.Is(err, ogenerrors.ErrSecurityRequirementIsNotSatisfied):
		return unauthenticated("Bearer")
	case errors.Is(err, domain.ErrUnauthenticated):
		return unauthenticated(`Bearer error="invalid_token"`)
	case errors.Is(err, domain.ErrNotFound):
		return problem(http.StatusNotFound, oas.ProblemCodeNotFound, "")
	case errors.Is(err, curtailment.ErrNoData):
		// The first files take a few minutes to load after a fresh start.
		p := problem(http.StatusServiceUnavailable, oas.ProblemCodeDataUnavailable, "The ONS's data is still being loaded.")
		p.RetryAfter = oas.NewOptInt(60)
		return p
	}

	if errors.Is(err, context.Canceled) {
		// The caller left; nobody reads this answer, and it is not our fault.
		h.logger.DebugContext(ctx, "request abandoned by the client", slog.Any("error", err))
	} else {
		h.logger.ErrorContext(ctx, "request failed", slog.Any("error", err))
	}
	return problem(http.StatusInternalServerError, oas.ProblemCodeInternal, "")
}

// handleRequestError answers a request the generated decoder refused, before
// any handler ran: a body that breaks the contract, malformed JSON, a wrong
// content type, a body over the size limit.
func (h *handler) handleRequestError(ctx context.Context, w http.ResponseWriter, _ *http.Request, err error) {
	var (
		fields      *validate.Error
		contentType *validate.InvalidContentTypeError
		tooLarge    *http.MaxBytesError
	)
	switch {
	case errors.As(err, &tooLarge):
		writeProblem(w, problem(http.StatusRequestEntityTooLarge, oas.ProblemCodeRequestTooLarge, "The body is larger than the API accepts."))
	case errors.As(err, &contentType):
		writeProblem(w, problem(http.StatusUnsupportedMediaType, oas.ProblemCodeUnsupportedMediaType, "Send the body as application/json."))
	case errors.As(err, &fields):
		p := problem(http.StatusUnprocessableEntity, oas.ProblemCodeValidationFailed, "Some fields break the rules.")
		for _, f := range fields.Fields {
			p.Response.Errors = append(p.Response.Errors, oas.FieldError{
				Pointer: "/" + strings.ReplaceAll(f.Name, ".", "/"),
				Code:    string(fieldCode(f.Error)),
			})
		}
		writeProblem(w, p)
	default:
		var (
			body   *ogenerrors.DecodeRequestError
			params *ogenerrors.DecodeParamsError
		)
		switch {
		case errors.As(err, &body):
			writeProblem(w, problem(http.StatusBadRequest, oas.ProblemCodeMalformedRequest, "The body is not what this operation reads."))
		case errors.As(err, &params):
			writeProblem(w, problem(http.StatusBadRequest, oas.ProblemCodeMalformedRequest, "A parameter is not what this operation reads."))
		default:
			h.logger.ErrorContext(ctx, "request failed outside a handler", slog.Any("error", err))
			writeProblem(w, problem(http.StatusInternalServerError, oas.ProblemCodeInternal, ""))
		}
	}
}

// fieldCode names the rule a field broke in the generated validation, in the
// words the domain uses for the same rule. ogen types the length errors; a
// number out of its range comes as text, which a test pins down so an ogen
// upgrade that rewords it fails the build instead of the form.
func fieldCode(err error) domain.Code {
	var (
		short *validate.MinLengthError
		long  *validate.MaxLengthError
	)
	switch {
	case errors.Is(err, validate.ErrFieldRequired):
		return domain.CodeRequired
	case errors.As(err, &short) && short.Len == 0:
		return domain.CodeRequired
	case errors.As(err, &short):
		return domain.CodeTooShort
	case errors.As(err, &long):
		return domain.CodeTooLong
	case strings.Contains(err.Error(), "less than") || strings.Contains(err.Error(), "greater than"):
		return domain.CodeOutOfRange
	default:
		return domain.CodeInvalid
	}
}

func (h *handler) notFound(w http.ResponseWriter, _ *http.Request) {
	writeProblem(w, problem(http.StatusNotFound, oas.ProblemCodeNotFound, "No operation lives at this path."))
}

func (h *handler) methodNotAllowed(w http.ResponseWriter, _ *http.Request, allowed string) {
	w.Header().Set("Allow", allowed)
	writeProblem(w, problem(http.StatusMethodNotAllowed, oas.ProblemCodeMethodNotAllowed, ""))
}

// unauthenticated answers 401 with the challenge RFC 9110 requires of it. In
// the words of RFC 6750, a request with no session gets the bare scheme and
// one with a session the API refuses gets invalid_token.
func unauthenticated(challenge string) *oas.ProblemStatusCodeWithHeaders {
	p := problem(http.StatusUnauthorized, oas.ProblemCodeUnauthenticated, "The session is missing, over or unknown.")
	p.WWWAuthenticate = oas.NewOptString(challenge)
	return p
}

func problem(status int, code oas.ProblemCode, detail string) *oas.ProblemStatusCodeWithHeaders {
	p := &oas.ProblemStatusCodeWithHeaders{
		StatusCode: status,
		Response: oas.Problem{
			Type:   "about:blank",
			Title:  http.StatusText(status),
			Status: status,
			Code:   code,
		},
	}
	if detail != "" {
		p.Response.Detail = oas.NewOptString(detail)
	}
	return p
}

// writeProblem writes a problem outside the generated encoder, for the
// answers ogen leaves to us: refused requests, unknown paths, panics.
func writeProblem(w http.ResponseWriter, p *oas.ProblemStatusCodeWithHeaders) {
	e := jx.GetEncoder()
	defer jx.PutEncoder(e)
	p.Response.Encode(e)

	w.Header().Set("Content-Type", "application/problem+json")
	if v, ok := p.RetryAfter.Get(); ok {
		w.Header().Set("Retry-After", strconv.Itoa(v))
	}
	if v, ok := p.WWWAuthenticate.Get(); ok {
		w.Header().Set("WWW-Authenticate", v)
	}
	w.WriteHeader(p.StatusCode)
	_, _ = w.Write(e.Bytes())
}
