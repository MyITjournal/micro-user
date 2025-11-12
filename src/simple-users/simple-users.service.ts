import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository, In } from 'typeorm';
import { SimpleUser } from './entity/simple-user.entity';
import {
  CreateSimpleUserInput,
  SimpleUserResponse,
  SimpleUserPreferencesResponse,
  BatchGetSimpleUserPreferencesInput,
  BatchGetSimpleUserPreferencesResponse,
  UpdateLastNotificationInput,
} from './dto/simple-user.dto';
import * as bcrypt from 'bcrypt';

@Injectable()
export class SimpleUsersService {
  constructor(
    @InjectRepository(SimpleUser)
    private readonly simpleUserRepository: Repository<SimpleUser>,
  ) {}

  async createUser(input: CreateSimpleUserInput): Promise<SimpleUserResponse> {
    console.log('SimpleUsersService.createUser called with:', input);

    // Generate user ID
    const userId = `usr_${Math.random().toString(36).substring(2, 10)}`;

    // Hash password
    const saltRounds = 10;
    const hashedPassword = await bcrypt.hash(input.password, saltRounds);

    // Create simple user
    const user = this.simpleUserRepository.create({
      user_id: userId,
      name: input.name,
      email: input.email,
      password: hashedPassword,
      push_token: input.push_token,
      email_preference: input.preferences.email,
      push_preference: input.preferences.push,
    });

    await this.simpleUserRepository.save(user);

    // Return response (without password)
    return {
      user_id: user.user_id,
      name: user.name,
      email: user.email,
      push_token: user.push_token,
      preferences: {
        email: user.email_preference,
        push: user.push_preference,
      },
      created_at: user.created_at,
    };
  }

  async getUserPreferences(
    userId: string,
  ): Promise<SimpleUserPreferencesResponse> {
    const user = await this.simpleUserRepository.findOne({
      where: { user_id: userId },
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

    return {
      user_id: user.user_id,
      email: user.email,
      preferences: {
        email: user.email_preference,
        push: user.push_preference,
      },
      last_notification_email: user.last_notification_email,
      last_notification_push: user.last_notification_push,
      last_notification_id: user.last_notification_id,
      updated_at: user.updated_at,
    };
  }

  async batchGetUserPreferences(
    input: BatchGetSimpleUserPreferencesInput,
  ): Promise<BatchGetSimpleUserPreferencesResponse> {
    // Remove duplicates from user_ids
    const uniqueUserIds = [...new Set(input.user_ids)];

    // Fetch all users in one query
    const users = await this.simpleUserRepository.find({
      where: { user_id: In(uniqueUserIds) },
    });

    // Build response for found users
    const foundUserIds = new Set(users.map((u) => u.user_id));
    const usersResponse: SimpleUserPreferencesResponse[] = users.map(
      (user) => ({
        user_id: user.user_id,
        email: user.email,
        preferences: {
          email: user.email_preference,
          push: user.push_preference,
        },
        last_notification_email: user.last_notification_email,
        last_notification_push: user.last_notification_push,
        last_notification_id: user.last_notification_id,
        updated_at: user.updated_at,
      }),
    );

    // Find not found user IDs
    const notFound = uniqueUserIds.filter((id) => !foundUserIds.has(id));

    return {
      users: usersResponse,
      not_found: notFound,
      total_requested: uniqueUserIds.length,
      total_found: usersResponse.length,
    };
  }

  async updateLastNotificationTime(
    userId: string,
    input: UpdateLastNotificationInput,
  ): Promise<void> {
    const user = await this.simpleUserRepository.findOne({
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
    this.simpleUserRepository.save(user).catch((error) => {
      // Log error but don't throw (fire-and-forget)
      console.error(
        `Failed to update last notification time for user ${userId}:`,
        error,
      );
    });
  }

  async getAllUsers(): Promise<SimpleUserPreferencesResponse[]> {
    const users = await this.simpleUserRepository.find({
      order: {
        created_at: 'DESC',
      },
    });

    return users.map((user) => ({
      user_id: user.user_id,
      email: user.email,
      preferences: {
        email: user.email_preference,
        push: user.push_preference,
      },
      last_notification_email: user.last_notification_email,
      last_notification_push: user.last_notification_push,
      last_notification_id: user.last_notification_id,
      updated_at: user.updated_at,
    }));
  }
}
