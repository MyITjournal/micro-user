import { Entity, Column, PrimaryGeneratedColumn, ManyToOne, OneToMany, JoinColumn, CreateDateColumn, UpdateDateColumn } from 'typeorm';
import { ObjectType, Field, ID } from '@nestjs/graphql';
import { User } from './user.entity';
import { UserDevice } from './userDevices.entity';

@ObjectType()
@Entity('user_channels')
export class UserChannel {
  @Field(() => ID)
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Field()
  @Column({ type: 'varchar', length: 50 })
  user_id: string;

  @Field()
  @Column({ type: 'varchar', length: 20 })
  channel_type: string; // 'email' or 'push'

  @Field()
  @Column({ type: 'boolean', default: true })
  enabled: boolean;

  @Field({ nullable: true })
  @Column({ type: 'boolean', default: false, nullable: true })
  verified?: boolean;

  @Field({ nullable: true })
  @Column({ type: 'varchar', length: 20, nullable: true })
  frequency?: string; // 'immediate', 'batched', 'digest'

  @Field({ nullable: true })
  @Column({ type: 'boolean', default: false, nullable: true })
  quiet_hours_enabled?: boolean;

  @Field({ nullable: true })
  @Column({ type: 'varchar', length: 5, nullable: true })
  quiet_hours_start?: string;

  @Field({ nullable: true })
  @Column({ type: 'varchar', length: 5, nullable: true })
  quiet_hours_end?: string;

  @Field({ nullable: true })
  @Column({ type: 'varchar', length: 50, nullable: true })
  quiet_hours_timezone?: string;

  @Field(() => User)
  @ManyToOne(() => User, user => user.channels)
  @JoinColumn({ name: 'user_id' })
  user: User;

  @Field(() => [UserDevice], { nullable: true })
  @OneToMany(() => UserDevice, device => device.channel, { cascade: true })
  devices?: UserDevice[];

  @Field()
  @CreateDateColumn()
  created_at: Date;

  @Field()
  @UpdateDateColumn()
  updated_at: Date;
}
