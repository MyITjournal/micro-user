import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository, In } from 'typeorm';
import { User } from './entity/user.entity';
import { UserChannel } from './entity/usersChannel.entity';
import { UserDevice } from './entity/userDevices.entity';
import {
  UserPreferencesResponse,
  ChannelsDto,
  CreateUserPreferencesInput,
  BatchGetUserPreferencesInput,
  BatchGetUserPreferencesResponse,
  OptOutStatusResponse,
  UpdateLastNotificationInput,
} from './dto/user.dto';

@Injectable()
export class UserService {
  constructor(
    @InjectRepository(User)
    private readonly userRepository: Repository<User>,
    @InjectRepository(UserChannel)
    private readonly channelRepository: Repository<UserChannel>,
    @InjectRepository(UserDevice)
    private readonly deviceRepository: Repository<UserDevice>,
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

  async submitUserPreferences(
    input: CreateUserPreferencesInput,
  ): Promise<UserPreferencesResponse> {
    // Check if user exists
    let user = await this.userRepository.findOne({
      where: { user_id: input.user_id },
      relations: ['channels', 'channels.devices'],
    });

    if (user) {
      // Update existing user
      user.email = input.email;
      user.phone = input.phone;
      user.timezone = input.timezone;
      user.language = input.language;
      user.notification_enabled = input.notification_enabled;
      user.marketing = input.preferences.marketing;
      user.transactional = input.preferences.transactional;
      user.reminders = input.preferences.reminders;
      user.digest_enabled = input.preferences.digest.enabled;
      user.digest_frequency = input.preferences.digest.frequency;
      user.digest_time = input.preferences.digest.time;

      await this.userRepository.save(user);

      // Delete existing channels and devices
      if (user.channels && user.channels.length > 0) {
        for (const channel of user.channels) {
          if (channel.devices) {
            await this.deviceRepository.remove(channel.devices);
          }
        }
        await this.channelRepository.remove(user.channels);
      }
    } else {
      // Create new user
      user = this.userRepository.create({
        user_id: input.user_id,
        email: input.email,
        phone: input.phone,
        timezone: input.timezone,
        language: input.language,
        notification_enabled: input.notification_enabled,
        marketing: input.preferences.marketing,
        transactional: input.preferences.transactional,
        reminders: input.preferences.reminders,
        digest_enabled: input.preferences.digest.enabled,
        digest_frequency: input.preferences.digest.frequency,
        digest_time: input.preferences.digest.time,
      });

      await this.userRepository.save(user);
    }

    // Create channels if provided
    if (input.channels) {
      const channels: UserChannel[] = [];

      // Create email channel
      const emailChannel = this.channelRepository.create({
        user_id: user.user_id,
        channel_type: 'email',
        enabled: input.channels.email.enabled,
        verified: input.channels.email.verified,
        frequency: input.channels.email.frequency,
        quiet_hours_enabled: input.channels.email.quiet_hours.enabled,
        quiet_hours_start: input.channels.email.quiet_hours.start,
        quiet_hours_end: input.channels.email.quiet_hours.end,
        quiet_hours_timezone: input.channels.email.quiet_hours.timezone,
      });
      channels.push(emailChannel);

      // Create push channel
      const pushChannel = this.channelRepository.create({
        user_id: user.user_id,
        channel_type: 'push',
        enabled: input.channels.push.enabled,
        quiet_hours_enabled: input.channels.push.quiet_hours.enabled,
        quiet_hours_start: input.channels.push.quiet_hours.start,
        quiet_hours_end: input.channels.push.quiet_hours.end,
        quiet_hours_timezone: input.channels.push.quiet_hours.timezone,
      });
      channels.push(pushChannel);

      await this.channelRepository.save(channels);

      // Create devices for push channel
      if (
        input.channels.push.devices &&
        input.channels.push.devices.length > 0
      ) {
        const devices = input.channels.push.devices.map((deviceInput) =>
          this.deviceRepository.create({
            device_id: deviceInput.device_id,
            channel_id: pushChannel.id,
            platform: deviceInput.platform,
            token: deviceInput.token,
            active: deviceInput.active,
            last_seen: new Date(),
          }),
        );
        await this.deviceRepository.save(devices);
      }
    }

    // Return the created/updated preferences
    return this.getUserPreferences(user.user_id, true);
  }

  async batchGetUserPreferences(
    input: BatchGetUserPreferencesInput,
  ): Promise<BatchGetUserPreferencesResponse> {
    // Remove duplicates from user_ids
    const uniqueUserIds = [...new Set(input.user_ids)];

    // Fetch all users in one query
    const users = await this.userRepository.find({
      where: { user_id: In(uniqueUserIds) },
      relations: input.include_channels ? ['channels', 'channels.devices'] : [],
    });

    // Build response for found users
    const foundUserIds = new Set(users.map((u) => u.user_id));
    const usersResponse: UserPreferencesResponse[] = users.map((user) => {
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

      if (input.include_channels && user.channels) {
        response.channels = this.formatChannels(user.channels);
      }

      return response;
    });

    // Find not found user IDs
    const notFound = uniqueUserIds.filter((id) => !foundUserIds.has(id));

    return {
      users: usersResponse,
      not_found: notFound,
      total_requested: uniqueUserIds.length,
      total_found: usersResponse.length,
    };
  }

  async getOptOutStatus(userId: string): Promise<OptOutStatusResponse> {
    // Use select to only fetch necessary fields for performance
    const user = await this.userRepository.findOne({
      where: { user_id: userId },
      relations: ['channels'],
      select: ['user_id', 'notification_enabled'],
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

    // Check if user has opted out globally or per channel
    const emailChannel = user.channels?.find((c) => c.channel_type === 'email');
    const pushChannel = user.channels?.find((c) => c.channel_type === 'push');

    const emailOptedOut = !user.notification_enabled || !emailChannel?.enabled;
    const pushOptedOut = !user.notification_enabled || !pushChannel?.enabled;
    const globalOptedOut = !user.notification_enabled;

    return {
      user_id: user.user_id,
      opted_out: globalOptedOut,
      channels: {
        email: emailOptedOut,
        push: pushOptedOut,
      },
      checked_at: new Date(),
    };
  }

  async updateLastNotificationTime(
    userId: string,
    input: UpdateLastNotificationInput,
  ): Promise<void> {
    const user = await this.userRepository.findOne({
      where: { user_id: userId },
    });

    if (!user) {
      // Silently ignore if user doesn't exist (fire-and-forget)
      return;
    }

    // Update the appropriate last notification field
    if (input.channel === 'email') {
      user.last_notification_email = new Date(input.sent_at);
    } else if (input.channel === 'push') {
      user.last_notification_push = new Date(input.sent_at);
    }

    user.last_notification_id = input.notification_id;

    // Save without waiting (fire-and-forget optimization)
    this.userRepository.save(user).catch((error) => {
      // Log error but don't throw (fire-and-forget)
      console.error(
        `Failed to update last notification time for user ${userId}:`,
        error,
      );
    });
  }
}
