package repository

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/github"
)

type manager struct {
	ghClient *github.Client
}

func New(ghClient *github.Client) *manager {
	return &manager{
		ghClient: ghClient,
	}
}

func (m *manager) CreateRepository(ctx context.Context, name string, fromRepository string) error {
	origin, err := m.fetchOrigin(ctx, fromRepository)
	if err != nil {
		return err
	}

	var owner string

	s := strings.Split(name, "/")
	if len(s) > 1 {
		owner, name = s[0], s[1]
	}
	user, _, err := m.ghClient.Users.Get(ctx, "")
	if err != nil {
		return err
	}

	var org string
	if *user.Login != owner {
		org = owner
	}

	init := true
	repo, _, err := m.ghClient.Repositories.Create(ctx, org, &github.Repository{
		Name:             &name,
		Description:      origin.repo.Description,
		Private:          origin.repo.Private,
		HasIssues:        origin.repo.HasIssues,
		HasDownloads:     origin.repo.HasDownloads,
		AllowRebaseMerge: origin.repo.AllowRebaseMerge,
		AllowSquashMerge: origin.repo.AllowSquashMerge,
		AllowMergeCommit: origin.repo.AllowMergeCommit,
		HasWiki:          origin.repo.HasWiki,
		HasPages:         origin.repo.HasPages,
		Homepage:         origin.repo.Homepage,
		AutoInit:         &init,
	})

	if err != nil {
		return err
	}

	err = m.initialCommit(ctx, origin.repo, repo)
	if err != nil {
		return fmt.Errorf("failed to clone origin repository: %w", err)
	}

	/*for _, team := range origin.teams {
		permissions := *collaborator.Permission

		_, err = m.ghClient.Repositories.Tea(ctx, owner, name, *collaborator.Login, &github.RepositoryAddCollaboratorOptions{
			Permission: *team.Permission,
		})

		if err != nil {
			return err
		}
	}*/

	_, _, err = m.ghClient.Repositories.ReplaceAllTopics(ctx, owner, name, origin.topics)

	if err != nil {
		return err
	}

	for _, branchProtection := range origin.branchProtecions {
		var restrictions *github.BranchRestrictionsRequest
		if branchProtection.protection.Restrictions != nil && (len(branchProtection.protection.Restrictions.Users) > 0 || len(branchProtection.protection.Restrictions.Teams) > 0) {
			restrictions = &github.BranchRestrictionsRequest{
				Users: usersToString(branchProtection.protection.Restrictions.Users),
				Teams: teamsToString(branchProtection.protection.Restrictions.Teams),
			}
		}

		var dismiss *github.DismissalRestrictionsRequest
		if branchProtection.protection.RequiredPullRequestReviews != nil && (len(branchProtection.protection.RequiredPullRequestReviews.DismissalRestrictions.Users) > 0 || len(branchProtection.protection.RequiredPullRequestReviews.DismissalRestrictions.Teams) > 0) {
			dismissalUsers := usersToString(branchProtection.protection.RequiredPullRequestReviews.DismissalRestrictions.Users)
			dismissalTeams := teamsToString(branchProtection.protection.RequiredPullRequestReviews.DismissalRestrictions.Teams)

			dismiss = &github.DismissalRestrictionsRequest{
				Users: &dismissalUsers,
				Teams: &dismissalTeams,
			}
		}

		_, _, err := m.ghClient.Repositories.UpdateBranchProtection(ctx, owner, name, *branchProtection.branch.Name, &github.ProtectionRequest{
			RequiredStatusChecks: branchProtection.protection.RequiredStatusChecks,
			RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
				DismissalRestrictionsRequest: dismiss,
				DismissStaleReviews:          branchProtection.protection.RequiredPullRequestReviews.DismissStaleReviews,
				RequireCodeOwnerReviews:      branchProtection.protection.RequiredPullRequestReviews.RequireCodeOwnerReviews,
				RequiredApprovingReviewCount: branchProtection.protection.RequiredPullRequestReviews.RequiredApprovingReviewCount,
			},
			EnforceAdmins: branchProtection.protection.EnforceAdmins.Enabled,
			Restrictions:  restrictions,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (m *manager) initialCommit(ctx context.Context, originRepo *github.Repository, bootstrapRepo *github.Repository) error {
	originDir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		return err
	}

	defer os.RemoveAll(originDir)

	_, err = git.PlainClone(originDir, false, &git.CloneOptions{
		URL:      *originRepo.SSHURL,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone origin repository: %w", err)
	}

	bootstrap, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		return err
	}

	if err := os.RemoveAll(path.Join(originDir, ".git")); err != nil {
		return err
	}

	defer os.RemoveAll(bootstrap)

	repo, err := git.PlainClone(bootstrap, false, &git.CloneOptions{
		URL:      *bootstrapRepo.SSHURL,
		Progress: os.Stdout,
	})

	if err != nil {
		return fmt.Errorf("failed to clone bootstrap repository: %w", err)
	}

	files, err := os.ReadDir(originDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		err := os.Rename(path.Join(originDir, file.Name()), path.Join(bootstrap, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to move from origin repository: %w", err)
		}
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed get git workdir: %w", err)
	}

	if err := w.AddGlob("*"); err != nil {
		return fmt.Errorf("failed to add files to repository workdir: %w", err)
	}

	if _, err := w.Commit(fmt.Sprintf("clone from origin repository %s/%s", *originRepo.Owner, *originRepo.Name), &git.CommitOptions{}); err != nil {
		return fmt.Errorf("failed commit initial commit: %w", err)
	}

	if err := repo.Push(&git.PushOptions{}); err != nil {
		return fmt.Errorf("failed to push initial commit: %w", err)
	}

	return nil
}

func usersToString(users []*github.User) []string {
	var result []string
	for _, u := range users {
		result = append(result, *u.Login)
	}

	return result
}

func teamsToString(teams []*github.Team) []string {
	var result []string
	for _, t := range teams {
		result = append(result, *t.Name)
	}

	return result
}

type origin struct {
	repo             *github.Repository
	topics           []string
	teams            []*github.Team
	branchProtecions []branchProtection
}

type branchProtection struct {
	branch     *github.Branch
	protection *github.Protection
}

func (m *manager) fetchOrigin(ctx context.Context, name string) (origin, error) {
	var owner string
	s := strings.Split(name, "/")
	if len(s) > 1 {
		owner, name = s[0], s[1]
	}

	origin := origin{}
	repo, _, err := m.ghClient.Repositories.Get(ctx, owner, name)
	if err != nil {
		return origin, fmt.Errorf("could not get origin repository: %w", err)
	}

	origin.repo = repo

	branches, _, err := m.ghClient.Repositories.ListBranches(ctx, owner, name, nil)
	if err != nil {
		return origin, err
	}

	for _, branch := range branches {
		protection, _, err := m.ghClient.Repositories.GetBranchProtection(ctx, owner, name, *branch.Name)
		if err != nil {
			continue
			//	return origin, err
		}

		if protection == nil {
			continue
		}

		origin.branchProtecions = append(origin.branchProtecions, branchProtection{
			branch:     branch,
			protection: protection,
		})
	}

	teams, _, err := m.ghClient.Repositories.ListTeams(ctx, owner, name, nil)
	if err != nil {
		return origin, err
	}

	origin.teams = teams

	topics, _, err := m.ghClient.Repositories.ListAllTopics(ctx, owner, name)
	if err != nil {
		return origin, err
	}

	origin.topics = topics

	return origin, nil
}
