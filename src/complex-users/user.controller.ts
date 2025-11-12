import {
  Controller,
  Get,
  Post,
  Body,
  Param,
  Query,
  HttpException,
  HttpStatus,
  HttpCode,
} from '@nestjs/common';
import { UserService } from './user.service';
import {
  UserPreferencesResponse,
  CreateUserPreferencesInput,
  BatchGetUserPreferencesInput,
  BatchGetUserPreferencesResponse,
  OptOutStatusResponse,
  UpdateLastNotificationInput,
} from './dto/user.dto';

@Controller('api/v1/cusers')
export class UserController {
  constructor(private readonly userService: UserService) {}

  @Get(':user_id/preferences')
  async getUserPreferences(
    @Param('user_id') userId: string,
    @Query('include_channels') includeChannels?: string,
  ): Promise<UserPreferencesResponse> {
    try {
      const includeChannelsBool =
        includeChannels === undefined ? true : includeChannels === 'true';

      return await this.userService.getUserPreferences(
        userId,
        includeChannelsBool,
      );
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }

      // Handle NotFoundException from service
      if (error.status === 404 || error.message?.includes('USER_NOT_FOUND')) {
        throw new HttpException(
          {
            error: {
              code: 'USER_NOT_FOUND',
              message: `User with ID ${userId} does not exist`,
              details: {
                user_id: userId,
              },
            },
          },
          HttpStatus.NOT_FOUND,
        );
      }

      // Handle other errors
      throw new HttpException(
        {
          error: {
            code: 'SERVICE_UNAVAILABLE',
            message: 'Unable to process request',
            details: {
              error: error.message,
            },
          },
        },
        HttpStatus.SERVICE_UNAVAILABLE,
      );
    }
  }

  @Post('preferences')
  async submitUserPreferences(
    @Body() input: CreateUserPreferencesInput,
  ): Promise<UserPreferencesResponse> {
    try {
      return await this.userService.submitUserPreferences(input);
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }

      // Handle validation or other errors
      throw new HttpException(
        {
          error: {
            code: 'BAD_REQUEST',
            message: 'Invalid input data',
            details: {
              error: error.message,
            },
          },
        },
        HttpStatus.BAD_REQUEST,
      );
    }
  }

  @Post('preferences/batch')
  async batchGetUserPreferences(
    @Body() input: BatchGetUserPreferencesInput,
  ): Promise<BatchGetUserPreferencesResponse> {
    try {
      // Validate max 100 users
      if (input.user_ids.length > 100) {
        throw new HttpException(
          {
            error: {
              code: 'VALIDATION_ERROR',
              message: 'Maximum 100 users allowed per batch request',
              details: {
                max_allowed: 100,
                requested: input.user_ids.length,
              },
            },
          },
          HttpStatus.BAD_REQUEST,
        );
      }

      // Check for duplicates
      const uniqueIds = new Set(input.user_ids);
      if (uniqueIds.size !== input.user_ids.length) {
        throw new HttpException(
          {
            error: {
              code: 'VALIDATION_ERROR',
              message: 'Duplicate user IDs are not allowed',
              details: {
                total: input.user_ids.length,
                unique: uniqueIds.size,
              },
            },
          },
          HttpStatus.BAD_REQUEST,
        );
      }

      return await this.userService.batchGetUserPreferences(input);
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }

      throw new HttpException(
        {
          error: {
            code: 'SERVICE_UNAVAILABLE',
            message: 'Unable to process batch request',
            details: {
              error: error.message,
            },
          },
        },
        HttpStatus.SERVICE_UNAVAILABLE,
      );
    }
  }

  @Get(':user_id/opt-out-status')
  async getOptOutStatus(
    @Param('user_id') userId: string,
  ): Promise<OptOutStatusResponse> {
    try {
      return await this.userService.getOptOutStatus(userId);
    } catch (error) {
      if (error instanceof HttpException) {
        throw error;
      }

      if (error.status === 404 || error.message?.includes('USER_NOT_FOUND')) {
        throw new HttpException(
          {
            error: {
              code: 'USER_NOT_FOUND',
              message: `User with ID ${userId} does not exist`,
              details: {
                user_id: userId,
              },
            },
          },
          HttpStatus.NOT_FOUND,
        );
      }

      throw new HttpException(
        {
          error: {
            code: 'SERVICE_UNAVAILABLE',
            message: 'Unable to check opt-out status',
            details: {
              error: error.message,
            },
          },
        },
        HttpStatus.SERVICE_UNAVAILABLE,
      );
    }
  }

  @Post(':user_id/last-notification')
  @HttpCode(204)
  async updateLastNotificationTime(
    @Param('user_id') userId: string,
    @Body() input: UpdateLastNotificationInput,
  ): Promise<void> {
    // Fire-and-forget - don't wait for response
    this.userService.updateLastNotificationTime(userId, input);
    return;
  }
}
