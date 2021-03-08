package testutils

import (
	"github.com/traPtitech/traQ/repository"
)

type EmptyTestRepository struct {
	repository.UserRepository
	repository.UserGroupRepository
	repository.TagRepository
	repository.ChannelRepository
	repository.MessageRepository
	repository.MessageReportRepository
	repository.StampRepository
	repository.StampPaletteRepository
	repository.StarRepository
	repository.PinRepository
	repository.DeviceRepository
	repository.FileRepository
	repository.WebhookRepository
	repository.OAuth2Repository
	repository.BotRepository
	repository.ClipRepository
	repository.OgpCacheRepository
	repository.UserSettingsRepository
}

func (*EmptyTestRepository) Sync() (init bool, err error) {
	return false, nil
}
