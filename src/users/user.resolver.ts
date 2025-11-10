import { Resolver, Query, Args } from '@nestjs/graphql';
import { UserService } from './user.service';
import {
  UserPreferencesResponse,
  GetUserPreferencesArgs,
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
}
