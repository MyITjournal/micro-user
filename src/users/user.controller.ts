import {
  Controller,
  Get,
  Post,
  Body,
  Param,
  Query,
  HttpException,
  HttpStatus,
} from '@nestjs/common';
import { UserService } from './user.service';
import {
  UserPreferencesResponse,
  CreateUserPreferencesInput,
} from './dto/user.dto';

@Controller('api/v1/users')
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
}
