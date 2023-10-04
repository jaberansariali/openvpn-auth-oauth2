package generic

import (
	"context"
	"fmt"
	"slices"

	"github.com/jkroepke/openvpn-auth-oauth2/internal/state"
	"github.com/jkroepke/openvpn-auth-oauth2/internal/types"
	"github.com/jkroepke/openvpn-auth-oauth2/internal/utils"
	"github.com/zitadel/oidc/v2/pkg/oidc"
)

func (p *Provider) CheckUser(
	_ context.Context,
	session state.State,
	_ types.UserData,
	tokens *oidc.Tokens[*oidc.IDTokenClaims],
) error {
	if err := p.CheckGroups(tokens); err != nil {
		return err
	}

	if err := p.CheckRoles(tokens); err != nil {
		return err
	}

	if err := p.CheckCommonName(session, tokens); err != nil {
		return err
	}

	return p.CheckIPAddress(session, tokens)
}

func (p *Provider) CheckGroups(tokens *oidc.Tokens[*oidc.IDTokenClaims]) error {
	if len(p.Conf.OAuth2.Validate.Groups) == 0 {
		return nil
	}

	tokenGroups, ok := tokens.IDTokenClaims.Claims["groups"]
	if !ok {
		return fmt.Errorf("%w: groups", ErrMissingClaim)
	}

	tokenGroupsList, err := utils.CastToSlice[string](tokenGroups)
	if err != nil {
		return fmt.Errorf("unable to decode token groups: %w", err)
	}

	for _, group := range p.Conf.OAuth2.Validate.Groups {
		if !slices.Contains(tokenGroupsList, group) {
			return fmt.Errorf("%w: %s", ErrMissingRequiredGroup, group)
		}
	}

	return nil
}

func (p *Provider) CheckRoles(tokens *oidc.Tokens[*oidc.IDTokenClaims]) error {
	if len(p.Conf.OAuth2.Validate.Roles) == 0 {
		return nil
	}

	tokenRoles, ok := tokens.IDTokenClaims.Claims["roles"]
	if !ok {
		return fmt.Errorf("%w: roles", ErrMissingClaim)
	}

	tokenRolesList, err := utils.CastToSlice[string](tokenRoles)
	if err != nil {
		return fmt.Errorf("unable to decode token roles: %w", err)
	}

	for _, role := range p.Conf.OAuth2.Validate.Roles {
		if !slices.Contains(tokenRolesList, role) {
			return fmt.Errorf("%w: %s", ErrMissingRequiredRole, role)
		}
	}

	return nil
}

func (p *Provider) CheckCommonName(session state.State, tokens *oidc.Tokens[*oidc.IDTokenClaims]) error {
	if p.Conf.OAuth2.Validate.CommonName == "" {
		return nil
	}

	tokenCommonName, ok := tokens.IDTokenClaims.Claims[p.Conf.OAuth2.Validate.CommonName].(string)
	if !ok {
		return fmt.Errorf("%w: %s", ErrMissingClaim, p.Conf.OAuth2.Validate.CommonName)
	}

	if tokenCommonName != session.CommonName {
		return fmt.Errorf("common_name %w: openvpn client: %s - oidc token: %s",
			ErrMismatch, tokenCommonName, session.CommonName)
	}

	return nil
}

func (p *Provider) CheckIPAddress(session state.State, tokens *oidc.Tokens[*oidc.IDTokenClaims]) error {
	if !p.Conf.OAuth2.Validate.IPAddr {
		return nil
	}

	tokenIpaddr, ok := tokens.IDTokenClaims.Claims["ipaddr"].(string)
	if !ok {
		return fmt.Errorf("%w: ipaddr", ErrMissingClaim)
	}

	if tokenIpaddr != session.Ipaddr {
		return fmt.Errorf("ipaddr %w: openvpn client: %s - oidc token: %s",
			ErrMismatch, tokenIpaddr, session.Ipaddr)
	}

	return nil
}