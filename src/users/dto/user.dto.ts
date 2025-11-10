import { ObjectType, Field, InputType, ArgsType } from '@nestjs/graphql';
import { IsBoolean, IsOptional } from 'class-validator';

// Query Arguments
@ArgsType()
export class GetUserPreferencesArgs {
  @Field(() => Boolean, { nullable: true, defaultValue: true })
  @IsOptional()
  @IsBoolean()
  include_channels?: boolean = true;
}

// Nested DTOs for Response
@ObjectType()
export class QuietHoursDto {
  @Field()
  enabled: boolean;

  @Field({ nullable: true })
  start?: string;

  @Field({ nullable: true })
  end?: string;

  @Field({ nullable: true })
  timezone?: string;
}

@ObjectType()
export class DeviceDto {
  @Field()
  device_id: string;

  @Field()
  platform: string;

  @Field()
  token: string;

  @Field()
  last_seen: Date;

  @Field()
  active: boolean;
}

@ObjectType()
export class EmailChannelDto {
  @Field()
  enabled: boolean;

  @Field()
  verified: boolean;

  @Field()
  frequency: string;

  @Field(() => QuietHoursDto)
  quiet_hours: QuietHoursDto;
}

@ObjectType()
export class PushChannelDto {
  @Field()
  enabled: boolean;

  @Field(() => [DeviceDto])
  devices: DeviceDto[];

  @Field(() => QuietHoursDto)
  quiet_hours: QuietHoursDto;
}

@ObjectType()
export class ChannelsDto {
  @Field(() => EmailChannelDto)
  email: EmailChannelDto;

  @Field(() => PushChannelDto)
  push: PushChannelDto;
}

@ObjectType()
export class DigestPreferenceDto {
  @Field()
  enabled: boolean;

  @Field()
  frequency: string;

  @Field()
  time: string;
}

@ObjectType()
export class UserPreferencesDto {
  @Field()
  marketing: boolean;

  @Field()
  transactional: boolean;

  @Field()
  reminders: boolean;

  @Field(() => DigestPreferenceDto)
  digest: DigestPreferenceDto;
}

// Main Response DTO
@ObjectType()
export class UserPreferencesResponse {
  @Field()
  user_id: string;

  @Field()
  email: string;

  @Field({ nullable: true })
  phone?: string;

  @Field()
  timezone: string;

  @Field()
  language: string;

  @Field()
  notification_enabled: boolean;

  @Field(() => ChannelsDto, { nullable: true })
  channels?: ChannelsDto;

  @Field(() => UserPreferencesDto)
  preferences: UserPreferencesDto;

  @Field()
  updated_at: Date;
}

// Error DTOs
@ObjectType()
export class ErrorDetails {
  @Field({ nullable: true })
  user_id?: string;

  @Field({ nullable: true })
  retry_after?: number;
}

@ObjectType()
export class ErrorDto {
  @Field()
  code: string;

  @Field()
  message: string;

  @Field(() => ErrorDetails)
  details: ErrorDetails;

  @Field()
  request_id: string;
}

@ObjectType()
export class ErrorResponse {
  @Field(() => ErrorDto)
  error: ErrorDto;
}
