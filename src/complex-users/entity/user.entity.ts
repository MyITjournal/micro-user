import {
  Entity,
  Column,
  PrimaryColumn,
  OneToMany,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';
import { ObjectType, Field, ID } from '@nestjs/graphql';
import { UserChannel } from './usersChannel.entity';

@ObjectType()
@Entity('users')
export class User {
  @Field(() => ID)
  @PrimaryColumn({ type: 'varchar', length: 50 })
  user_id: string;

  @Field()
  @Column({ type: 'varchar', length: 255, default: 'User' })
  name: string;

  @Field()
  @Column({ type: 'varchar', length: 255, unique: true })
  email: string;

  @Column({ type: 'varchar', length: 255, default: '' })
  password: string;

  @Field({ nullable: true })
  @Column({ type: 'text', nullable: true })
  push_token?: string;

  @Field({ nullable: true })
  @Column({ type: 'varchar', length: 20, nullable: true })
  phone?: string;

  @Field()
  @Column({ type: 'varchar', length: 50, default: 'UTC' })
  timezone: string;

  @Field()
  @Column({ type: 'varchar', length: 10, default: 'en' })
  language: string;

  @Field()
  @Column({ type: 'boolean', default: true })
  notification_enabled: boolean;

  @Field()
  @Column({ type: 'boolean', default: true })
  email_preference: boolean;

  @Field()
  @Column({ type: 'boolean', default: true })
  push_preference: boolean;

  @Field()
  @Column({ type: 'boolean', default: false })
  marketing: boolean;

  @Field()
  @Column({ type: 'boolean', default: true })
  transactional: boolean;

  @Field()
  @Column({ type: 'boolean', default: true })
  reminders: boolean;

  @Field()
  @Column({ type: 'boolean', default: false })
  digest_enabled: boolean;

  @Field()
  @Column({ type: 'varchar', length: 20, default: 'daily' })
  digest_frequency: string;

  @Field()
  @Column({ type: 'varchar', length: 5, default: '09:00' })
  digest_time: string;

  @Field(() => [UserChannel], { nullable: true })
  @OneToMany(() => UserChannel, (channel) => channel.user, { cascade: true })
  channels: UserChannel[];
  @Field({ nullable: true })
  @Column({ type: 'timestamp', nullable: true })
  last_notification_email?: Date;

  @Field({ nullable: true })
  @Column({ type: 'timestamp', nullable: true })
  last_notification_push?: Date;

  @Field({ nullable: true })
  @Column({ type: 'varchar', length: 100, nullable: true })
  last_notification_id?: string;

  @Field()
  @CreateDateColumn()
  created_at: Date;

  @Field()
  @UpdateDateColumn()
  updated_at: Date;
}
