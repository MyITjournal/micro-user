import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { User } from './entity/user.entity';
import { UserChannel } from './entity/usersChannel.entity';
import {
  UserPreferencesResponse,
  ChannelsDto,
  EmailChannelDto,
  PushChannelDto,
  QuietHoursDto,
  DeviceDto,
  UserPreferencesDto,
  DigestPreferenceDto,
} from './dto/user.dto';

@Injectable()
export class UserService {
  constructor(
    @InjectRepository(User)
    private readonly userRepository: Repository<User>,
    @InjectRepository(UserChannel)
    private readonly channelRepository: Repository<UserChannel>,
  ) {}

  async getUserPreferences(
    userId: string,
    includeChannels: boolean = true,
  ): Promise<UserPreferencesResponse> {
    const user = await this.userRepository.findOne({
      where: { user_id: userId },
      relations: includeChannels ? ['channels', 'channels.devices'] : [],
    });

    if (!user) {
      throw new NotFoundException({
        code: 'USER_NOT_FOUND',
        message: `User with ID ${userId} does not exist`,
        details: {
          user_id: userId,
        },
      });
    }

    const response: UserPreferencesResponse = {
      user_id: user.user_id,
      email: user.email,
      phone: user.phone,
      timezone: user.timezone,
      language: user.language,
      notification_enabled: user.notification_enabled,
      preferences: {
        marketing: user.marketing,
        transactional: user.transactional,
        reminders: user.reminders,
        digest: {
          enabled: user.digest_enabled,
          frequency: user.digest_frequency,
          time: user.digest_time,
        },
      },
      updated_at: user.updated_at,
    };

    if (includeChannels && user.channels) {
      response.channels = this.formatChannels(user.channels);
    }

    return response;
  }

  private formatChannels(channels: UserChannel[]): ChannelsDto {
    const emailChannel = channels.find((c) => c.channel_type === 'email');
    const pushChannel = channels.find((c) => c.channel_type === 'push');

    const result: ChannelsDto = {
      email: {
        enabled: emailChannel?.enabled ?? false,
        verified: emailChannel?.verified ?? false,
        frequency: emailChannel?.frequency ?? 'immediate',
        quiet_hours: {
          enabled: emailChannel?.quiet_hours_enabled ?? false,
          start: emailChannel?.quiet_hours_start,
          end: emailChannel?.quiet_hours_end,
          timezone: emailChannel?.quiet_hours_timezone,
        },
      },
      push: {
        enabled: pushChannel?.enabled ?? false,
        devices:
          pushChannel?.devices?.map((device) => ({
            device_id: device.device_id,
            platform: device.platform,
            token: device.token,
            last_seen: device.last_seen,
            active: device.active,
          })) ?? [],
        quiet_hours: {
          enabled: pushChannel?.quiet_hours_enabled ?? false,
          start: pushChannel?.quiet_hours_start,
          end: pushChannel?.quiet_hours_end,
          timezone: pushChannel?.quiet_hours_timezone,
        },
      },
    };

    return result;
  }
}
