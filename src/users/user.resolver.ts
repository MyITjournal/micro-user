import { Resolver, Query, Mutation, Args } from '@nestjs/graphql';
import { UserService } from './user.service';
import {
  UserPreferencesResponse,
  GetUserPreferencesArgs,
  CreateUserPreferencesInput,
  BatchGetUserPreferencesInput,
  BatchGetUserPreferencesResponse,
  OptOutStatusResponse,
  UpdateLastNotificationInput,
} from './dto/user.dto';

@Resolver()
export class UserResolver {
  constructor(private readonly userService: UserService) {}

  @Query(() => UserPreferencesResponse, {
    name: 'getUserPreferences',
    description: 'Retrieve notification preferences for a specific user',
  })
  async getUserPreferences(
    @Args('user_id') userId: string,
    @Args() args: GetUserPreferencesArgs,
  ): Promise<UserPreferencesResponse> {
    return this.userService.getUserPreferences(userId, args.include_channels);
  }

  @Mutation(() => UserPreferencesResponse, {
    name: 'submitUserPreferences',
    description: 'Create or update user notification preferences',
  })
  async submitUserPreferences(
    @Args('input') input: CreateUserPreferencesInput,
  ): Promise<UserPreferencesResponse> {
    return this.userService.submitUserPreferences(input);
  }

  @Query(() => BatchGetUserPreferencesResponse, {
    name: 'batchGetUserPreferences',
    description: 'Retrieve preferences for multiple users in a single request',
  })
  async batchGetUserPreferences(
    @Args('input') input: BatchGetUserPreferencesInput,
  ): Promise<BatchGetUserPreferencesResponse> {
    return this.userService.batchGetUserPreferences(input);
  }

  @Query(() => OptOutStatusResponse, {
    name: 'getOptOutStatus',
    description: 'Quick check if user has opted out (lightweight endpoint)',
  })
  async getOptOutStatus(
    @Args('user_id') userId: string,
  ): Promise<OptOutStatusResponse> {
    return this.userService.getOptOutStatus(userId);
  }

  @Mutation(() => Boolean, {
    name: 'updateLastNotificationTime',
    description: 'Update when user last received a notification',
  })
  async updateLastNotificationTime(
    @Args('user_id') userId: string,
    @Args('input') input: UpdateLastNotificationInput,
  ): Promise<boolean> {
    await this.userService.updateLastNotificationTime(userId, input);
    return true;
  }
}
