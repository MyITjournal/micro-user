import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { UserResolver } from './user.resolver';
import { UserService } from './user.service';
import { User } from './entity/user.entity';
import { UserChannel } from './entity/usersChannel.entity';
import { UserDevice } from './entity/userDevices.entity';

@Module({
  imports: [TypeOrmModule.forFeature([User, UserChannel, UserDevice])],
  providers: [UserResolver, UserService],
  exports: [UserService],
})
export class UsersModule {}
