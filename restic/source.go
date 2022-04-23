package restic

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	sup "github.com/creativeprojects/go-selfupdate"
)

type cachedSource struct {
	owner, repo string
	source      sup.Source
	releases    []sup.SourceRelease
	lastUpdated time.Time
	lock        sync.Mutex
}

func (s *cachedSource) checkOwnerAndRepo(repository sup.Repository) error {
	o, r, _ := repository.GetSlug()
	if o != s.owner || r != s.repo {
		return fmt.Errorf("expected owner %q == %q && repo %q == %q", s.owner, o, s.repo, r)
	}
	return nil
}

func (s *cachedSource) ListReleases(ctx context.Context, repository sup.Repository) (releases []sup.SourceRelease, err error) {
	if err = s.checkOwnerAndRepo(repository); err != nil {
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	now := time.Now()
	if s.releases != nil && now.Sub(s.lastUpdated).Minutes() < 15 {
		releases = s.releases
	} else if releases, err = s.source.ListReleases(ctx, sup.NewRepositorySlug(s.owner, s.repo)); err == nil {
		s.releases = releases
		s.lastUpdated = now
	}

	return
}

func (s *cachedSource) Reset() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.releases = nil
}

func (s *cachedSource) DownloadReleaseAsset(ctx context.Context, rel *sup.Release, assetID int64) (content io.ReadCloser, err error) {
	return s.source.DownloadReleaseAsset(ctx, rel, assetID)
}

func newCachedSource(owner, repo string, source sup.Source) *cachedSource {
	return &cachedSource{
		owner:  owner,
		repo:   repo,
		source: source,
	}
}
