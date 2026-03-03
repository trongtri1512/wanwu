<template>
  <div class="layout full-menu" :style="`background: ${bgColor}`">
    <el-container class="outer-container">
      <!-- 左侧内容 -->
      <div
        v-if="isShowMenu"
        :class="[
          'left-aside-container',
          { 'left-aside-container-isCollapse': isCollapse },
        ]"
      >
        <div class="left-header-container">
          <div
            style="padding-top: 10px; text-align: center"
            v-if="homeLogoPath"
          >
            <img
              v-if="!isCollapse"
              style="height: 46px"
              :src="avatarSrc(homeLogoPath)"
            />
            <img
              v-else
              style="width: 50%; margin-top: 10px"
              src="@/assets/imgs/wanwu.png"
            />
          </div>
          <!-- 组织切换 -->
          <div class="menu-org-select-wrapper" v-if="!isCollapse">
            <img style="width: 16px" src="@/assets/imgs/org_user.svg" alt="" />
            <ChangeOrg
              :org="org"
              :orgList="orgList"
              :getCurrentOrgName="getCurrentOrgName"
              :changeOrg="changeOrg"
            />
          </div>
          <el-popover
            v-else
            popper-class="menu-org-popover"
            placement="right"
            width="220"
            trigger="click"
          >
            <ChangeOrg
              :org="org"
              :orgList="orgList"
              :getCurrentOrgName="getCurrentOrgName"
              :changeOrg="changeOrg"
            />
            <div slot="reference">
              <div style="text-align: center; cursor: pointer">
                <img
                  style="width: 20px; margin-top: 24px"
                  src="@/assets/imgs/org_user.svg"
                  alt=""
                />
              </div>
            </div>
          </el-popover>
        </div>
        <!-- 菜单 -->
        <el-aside
          v-if="menuList && menuList.length"
          :class="['full-menu-aside', { 'full-menu-isCollapse': isCollapse }]"
        >
          <el-menu
            :default-openeds="defaultOpeneds"
            :default-active="activeIndex"
            :key="menuKey"
            :collapse="isCollapse"
          >
            <!--菜单渲染-->
            <div v-for="(n, i) in menuList" :key="`${i}ml`">
              <!--有下一级-->
              <el-submenu
                v-if="n.children && checkPerm(n.perm)"
                :index="n.index"
              >
                <template slot="title">
                  <div class="menu-svg">
                    <svg-icon class="menu-icon" :icon-class="n.icon" />
                  </div>
                  <span class="menu-withIcon-title">{{ n.name }}</span>
                </template>
                <div
                  v-for="(m, j) in n.children"
                  v-if="checkPerm(m.perm)"
                  :key="`${j}cl`"
                >
                  <el-submenu
                    v-if="m.children"
                    :index="m.index"
                    :class="['menu-indent']"
                  >
                    <template slot="title">{{ m.name }}</template>
                    <div
                      v-for="(p, k) in m.children"
                      :key="`${k}pl`"
                      v-if="checkPerm(p.perm)"
                    >
                      <el-submenu
                        v-if="p.children"
                        :index="p.index"
                        :class="['menu-indent-sub']"
                      >
                        <template slot="title">{{ p.name }}</template>
                        <el-menu-item
                          v-for="(item, index) in p.children"
                          :key="`${index}itemEl`"
                          :index="item.index"
                          v-if="checkPerm(item.perm)"
                          @click="menuClick(item)"
                          :class="[{ 'is-active': activeIndex === item.index }]"
                        >
                          {{ item.name }}
                        </el-menu-item>
                      </el-submenu>
                      <el-menu-item
                        v-else
                        :index="p.index"
                        @click="menuClick(p)"
                        :class="[{ 'is-active': activeIndex === p.index }]"
                      >
                        {{ p.name }}
                      </el-menu-item>
                    </div>
                  </el-submenu>
                  <el-menu-item
                    v-else
                    :index="m.index"
                    @click="menuClick(m)"
                    :class="[
                      'menu-indent-item',
                      { 'is-active': activeIndex === m.index },
                    ]"
                  >
                    {{ m.name }}
                  </el-menu-item>
                </div>
              </el-submenu>
              <!--没有下一级-->
              <el-menu-item
                :index="n.index"
                v-if="!n.children && checkPerm(n.perm)"
                @click="menuClick(n)"
                :class="[{ 'is-active': activeIndex === n.index }]"
              >
                <div class="menu-svg">
                  <svg-icon class="menu-icon" :icon-class="n.icon" />
                </div>
                <span class="menu-withIcon-title">{{ n.name }}</span>
              </el-menu-item>
            </div>
          </el-menu>
        </el-aside>
        <div
          :class="['left-bottom-container', { 'menu-isCollapse': isCollapse }]"
        >
          <el-popover placement="top" width="220" trigger="click">
            <div
              :class="[
                'menu--popover-wrap',
                { 'wrap-last': popoverList.length === index + 1 },
              ]"
              v-for="(it, index) in popoverList"
              :key="'popoverList' + index"
            >
              <div
                v-if="checkPerm(item.perm)"
                v-for="item in it"
                :key="item.name"
                class="menu--popover-item"
                @click="menuClick(item)"
              >
                <img class="menu--popover-item-img" :src="item.img" alt="" />
                <el-tooltip
                  v-if="item.isTip"
                  effect="dark"
                  :content="item.tipContent"
                  placement="top-start"
                >
                  <span
                    style="display: inline-block; width: 150px"
                    class="menu--popover-item-name"
                  >
                    {{ item.name }}
                  </span>
                </el-tooltip>
                <span v-if="!item.isTip" class="menu--popover-item-name">
                  {{ item.name }}
                </span>
                <img
                  v-if="item.icon"
                  class="menu--popover-item-icon"
                  :src="item.icon"
                  alt=""
                />
                <span v-if="item.version" class="menu--popover-item-version">
                  {{ version || '' }}
                </span>
              </div>
            </div>
            <div
              v-if="!isCollapse"
              class="header-user-content"
              slot="reference"
            >
              <img
                class="header-icon"
                v-if="userAvatar"
                :src="avatarSrc(userAvatar)"
              />
              <div class="header-user-name">{{ userInfo.userName }}</div>
              <i class="el-icon-more header-more"></i>
            </div>
            <div class="user-content-isCollapse" v-else slot="reference">
              <img
                class="header-icon"
                v-if="userAvatar"
                :src="avatarSrc(userAvatar)"
              />
            </div>
          </el-popover>
          <div class="left-bottom-menu-icon" @click="changeMenuCollapse">
            <img
              src="@/assets/imgs/left_menu_icon.svg"
              alt=""
              v-if="!isCollapse"
            />
            <img src="@/assets/imgs/right_menu_icon.svg" alt="" v-else />
          </div>
        </div>
      </div>
      <!-- 容器 -->
      <el-container :class="['inner-container']">
        <!-- 右侧内容 -->
        <el-main :class="[{ 'no-header-main': !isShowMenu }]">
          <div class="page-container">
            <div class="right-page-content">
              <router-view></router-view>
              <div id="container" class="qk-container"></div>
            </div>
          </div>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script>
// import { start } from 'qiankun'
import { mapActions, mapGetters } from 'vuex';
import { checkPerm, PERMS } from '@/router/permission';
import { menuList } from './menu';
import { changeLang } from '@/api/user';
import {
  fetchPermFirPath,
  fetchCurrentPathIndex,
  replaceIcon,
  replaceTitle,
  redirectUserInfoPage,
  avatarSrc,
} from '@/utils/util';
import ChangeLang from '@/components/changeLang.vue';
import ChangeOrg from '@/components/changeOrg.vue';
import { DOC_FIRST_KEY } from '@/views/docCenter/constants';

export default {
  name: 'Layout',
  components: { ChangeLang, ChangeOrg },
  data() {
    return {
      isCollapse: false,
      homeLogoPath: '',
      bgColor: '',
      version: '',
      orgList: [],
      org: { orgId: '' },
      defaultOpeneds: [],
      menuList: [],
      menuKey: 'menu_key',
      activeIndex: '',
      isShowMenu: true,
      popoverList: [
        [
          {
            name: this.$t('menu.account'),
            path: '/userInfo',
            img: require('@/assets/imgs/user_icon.svg'),
          },
          {
            name: this.$t('menu.setting'),
            path: '/permission',
            img: require('@/assets/imgs/setting_icon.svg'),
            isTip: true,
            tipContent: this.$t('menu.settingTip'),
            perm: PERMS.PERMISSION,
          },
          {
            name: this.$t('menu.operationManage'),
            path: '/operation',
            img: require('@/assets/imgs/operationManage.svg'),
            perm: PERMS.OPERATION,
          },
        ],
        [
          {
            name: this.$t('menu.helpDoc'),
            img: require('@/assets/imgs/helpDoc_icon.svg'),
            icon: require('@/assets/imgs/link_icon.png'),
            redirect: () => {
              // window.open('https://github.com/UnicomAI/wanwu/tree/main/docs/manual')
              window.open(
                window.location.origin +
                  `${this.$basePath}/aibase/docCenter/pages/${DOC_FIRST_KEY}`,
              );
            },
          },
          {
            name: 'Github',
            img: require('@/assets/imgs/github_icon.svg'),
            icon: require('@/assets/imgs/link_icon.png'),
            redirect: () => {
              window.open('https://github.com/UnicomAI/wanwu');
            },
          },
          {
            name: this.$t('menu.about'),
            img: require('@/assets/imgs/about_icon.svg'),
            version: 'version',
          },
        ],
        [
          {
            name: this.$t('header.logout'),
            img: require('@/assets/imgs/logout_icon.svg'),
            redirect: () => {
              this.logout();
            },
          },
        ],
      ],
    };
  },
  watch: {
    $route: {
      handler(val) {
        this.justifyIsShowMenu(val.path);
        this.getMenuList(val.path);
        this.redirectUserInfo();
        this.initScroll();
      },
      // 深度观察监听
      deep: true,
    },
    orgInfo: {
      handler(val) {
        this.orgList = val.orgs || [];
      },
      deep: true,
    },
    commonInfo: {
      handler(val) {
        const { home = {}, tab = {}, about = {} } = val.data || {};
        this.homeLogoPath = home.logo ? home.logo.path : '';
        this.bgColor = home.backgroundColor || this.$config.backgroundColor;
        this.version = about.version || '1.0';
        replaceIcon(tab.logo ? tab.logo.path : '');
        replaceTitle(tab.title);
      },
      deep: true,
    },
    permission: {
      handler() {
        // 如果没修改过密码，重新向到修改密码
        this.redirectUserInfo();
      },
      deep: true,
    },
  },
  computed: {
    ...mapGetters('user', [
      'orgInfo',
      'userInfo',
      'commonInfo',
      'permission',
      'userAvatar',
    ]),
  },
  async created() {
    // 判断是否展示左侧菜单
    this.justifyIsShowMenu(this.$route.path);

    // 设置语言
    // await this.setLanguage()

    // 获取菜单
    this.getCurrentMenu();
    this.setLocalMenuCollapse();

    // 只有登陆状态下才查询接口，否则会一直刷新
    if (localStorage.getItem('access_cert')) this.getPermissionInfo();

    // 设置组织列表以及当前的组织
    this.orgList = this.orgInfo.orgs || [];
    this.org.orgId = this.userInfo.orgId;

    // 获取平台名称以及 logo 等信息
    this.getCommonInfo();
  },
  /* 保证容器 DIV 在 qiankun start 时一定存在 */
  mounted() {
    /* start() */
  },
  methods: {
    avatarSrc,
    ...mapActions('user', ['LoginOut', 'getPermissionInfo', 'getCommonInfo']),
    checkPerm,
    logout() {
      window.localStorage.removeItem('access_cert');
      window.location.href =
        window.location.origin + this.$basePath + '/aibase/login';
    },
    setLocalMenuCollapse() {
      this.isCollapse = localStorage.getItem('menu_collapse') === 'true';
    },
    initScroll() {
      const pageContainer = document.querySelector('.el-main');
      if (pageContainer) {
        pageContainer.scrollTop = 0;
        pageContainer.scrollLeft = 0;
      }
    },
    getCurrentOrgName() {
      const currentOrg =
        this.orgList.filter(item => item.id === this.org.orgId)[0] || {};
      return currentOrg.name;
    },
    redirectUserInfo() {
      redirectUserInfoPage(this.permission.isUpdatePassword, () => {
        return null;
      });
    },
    justifyDocPages(val) {
      const path = `${this.$basePath}/aibase` + val;
      return val && path.includes(`${this.$basePath}/aibase/docCenter/pages`);
    },
    justifyIsShowMenu(path) {
      const notShowArr = ['/workflow'];
      let isShowMenu = true;
      if (this.justifyDocPages(path)) {
        isShowMenu = false;
      } else {
        for (let item of notShowArr) {
          if (item === path) {
            isShowMenu = false;
            break;
          }
        }
      }
      this.isShowMenu = isShowMenu;
    },
    async setLanguage() {
      const langCode = localStorage.getItem('locale');
      // 主要解决本地和线上两个 localStorage 语言不同问题，使用用户本地缓存的语言
      if (langCode) await changeLang({ language: langCode });
    },
    setMenuOpeneds(menus) {
      this.defaultOpeneds = menus.map(item => item.index);
    },
    menuClick(item) {
      if (item.redirect) {
        item.redirect();
        this.changeMenuIndex(item.index);
      } else {
        if (item.path) this.$router.push({ path: item.path });
      }
    },
    getCurrentMenu() {
      const { path } = this.$route || {};
      // 获取当前菜单
      this.getMenuList(path);
    },
    getMenuList(path) {
      // 获取当前菜单列表
      const menus = menuList;

      this.menuList = menus;
      this.setMenuOpeneds(menus);

      // 给当前 activeIndex 赋值
      this.changeMenuIndex(fetchCurrentPathIndex(path, menus));
    },
    changeMenuIndex(index) {
      this.activeIndex = index;
    },
    async changeOrg(orgId) {
      this.$store.state.user.userInfo.orgId = orgId;
      // 切换组织更新权限，跳转有权限的页面；如果是用模型跳转用模型，其他跳转模型开发平台
      await this.getPermissionInfo();

      // 更新 storage 用户信息中组织 id
      const info = JSON.parse(localStorage.getItem('access_cert'));
      info.user.orgId = orgId;
      localStorage.setItem('access_cert', JSON.stringify(info));

      const { path } = fetchPermFirPath();
      // 如果当前页面 path 与第一个有权限的 path 相同，需要刷新页面以确保数据为新切换组织的
      if (path === this.$route.path) {
        location.reload();
        return;
      }
      // 切换组织, 根据当前路径有权限的第一个路径找到对应的 menu
      this.getMenuList(path);
      this.menuClick({ path });
    },
    changeMenuCollapse() {
      this.isCollapse = !this.isCollapse;
      if (!this.isCollapse) {
        this.setMenuOpeneds(this.menuList);
      }
      localStorage.setItem('menu_collapse', this.isCollapse);
    },
  },
};
</script>

<style lang="scss" scoped>
.disabled {
  cursor: not-allowed !important;
}
.full-menu.layout {
  height: 100%;
  .outer-container {
    height: 100%;

    .left-aside-container {
      position: relative;
      width: 208px;
      background: #fff;
      border-right: 1px solid #d8d8d8;
      transition: width 0.25s linear !important;
      .left-header-container {
        position: absolute;
        top: 0;
        width: 100%;
        z-index: 10;
      }
      .left-bottom-container {
        width: 100%;
        height: 68px;
        display: flex;
        align-items: center;
        justify-content: space-around;
        padding: 0 16px;
        .left-bottom-menu-icon {
          img {
            width: 18px;
            height: 18px;
            cursor: pointer;
            margin-top: 5px;
          }
        }
        .header-user-content {
          display: flex;
          align-items: center;
          padding: 0 16px;
          height: 40px;
          width: 142px;
          border-radius: 30px;
          background: #fff;
          color: #1f1f1f;
          box-shadow: 0 2px 8px 0 rgba(16, 18, 25, 0.102);
          font-size: 14px;
          cursor: pointer;
          margin-right: 16px;

          .header-user-name {
            width: 46px;
            overflow: hidden;
            white-space: nowrap;
            text-overflow: ellipsis;
          }

          .header-icon {
            width: 26px;
            height: 26px;
            border-radius: 50%;
            margin-right: 10px;
          }

          .header-more {
            margin-left: 8px;
            transform: rotate(90deg);
            font-size: 16px;
            color: #a2a7b4;
          }
        }
      }
      .left-bottom-container.menu-isCollapse {
        display: block;
        text-align: center;
        height: 90px;
        margin-top: 5px;
        .user-content-isCollapse {
          padding-bottom: 10px;
          border-bottom: 1px solid #d8d8d8;
          margin-bottom: 5px;
          cursor: pointer;
        }
        .header-icon {
          width: 28px;
          height: 28px;
          border-radius: 50%;
          margin: 0 auto;
        }
      }
      .menu-org-select-wrapper ::v-deep {
        display: flex;
        align-items: center;
        margin: 8px 14px 0;
        border: 1px solid #dcdfe6;
        border-radius: 4px;
        padding-left: 12px;
        .el-select .el-input.is-focus .el-input__inner,
        .el-input__inner,
        .el-input__inner:focus {
          border: none !important;
          outline: none !important;
          padding-left: 12px;
        }
      }
    }

    .left-aside-container.left-aside-container-isCollapse {
      width: 65px;
      .el-aside.full-menu-aside {
        height: calc(100vh - 200px) !important;
      }
    }

    /*element ui 样式重写*/
    .el-aside.full-menu-aside {
      height: calc(100vh - 178px);
      width: 100% !important;
      border-radius: 10px 0 0 10px;
      margin-top: 110px;
      position: relative;

      .el-menu {
        height: 100%;
        width: 100%;
        overflow-x: auto;
        overflow-y: auto;
        border: none;
        .menu-indent ::v-deep .el-submenu__title,
        .menu-indent-item {
          padding-left: 45px !important;
        }
        .menu-indent-sub ::v-deep .el-submenu__title {
          padding-left: 60px !important;
        }
        .menu-withIcon-title {
          display: inline-block;
        }
        .menu-svg {
          .menu-icon {
            font-size: 16px;
            margin-right: 10px;
          }
          .svg-icon {
            color: #868d9c;
          }
        }
      }
      .el-menu ::v-deep {
        .el-menu-item {
          color: $color_title;
        }
        .el-submenu__title,
        .el-menu-item span {
          font-size: 15px !important;
        }
        .el-menu-item.is-active,
        .el-menu-item:focus {
          background-color: $color_opacity !important;
        }
        .el-menu-item.is-active,
        .el-submenu.is-active {
          .el-submenu__title:hover {
            background-color: $color_opacity !important;
          }
          .svg-icon {
            color: $color !important;
          }
        }
        .el-submenu__title {
          span {
            font-size: 15px !important;
          }
        }
        .el-submenu.is-active .el-submenu__title {
          border-bottom-color: $color !important;
        }
        .el-submenu__title,
        .el-menu-item {
          height: 40px;
          display: flex;
          align-items: center;
          border-radius: 6px;
          margin: 6px;
          min-width: auto;
          font-size: 15px;
        }
      }
      .el-menu--collapse.el-menu ::v-deep {
        width: 64px;
        .el-submenu__title {
          padding-left: 0 !important;
          display: block !important;
        }
        .el-submenu__icon-arrow.el-icon-arrow-right,
        .menu-withIcon-title {
          display: none;
        }
        .menu-svg {
          width: 44px;
          height: 40px;
          border-radius: 6px;
          text-align: center;
          line-height: 40px;
          .menu-icon {
            font-size: 18px;
            margin-right: 0;
          }
        }
        .el-submenu.is-active,
        .el-menu-item.is-active {
          .menu-svg {
            background: linear-gradient(
              145deg,
              #57e4fd -1%,
              #41c9ff 32%,
              #4161fe 114%
            );
            .svg-icon {
              color: #fff !important;
            }
          }
        }
        .el-submenu__title,
        .el-menu-item {
          margin: 12px 10px;
        }
        .el-menu-item {
          padding: 0 !important;
        }
      }
    }

    .inner-container {
      width: calc(100% - 80px);
      height: 100%;
      border-radius: 10px;
      .el-header {
        line-height: 60px;
        height: 60px;
        box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
        z-index: 2;
        background: url('@/assets/imgs/nav_bg.png');
        background-size: 100% 100%;
      }
      .el-main {
        position: relative;
        margin: 0 !important;
        padding: 0 !important;
        width: 100%;
        height: 100%;
        overflow: auto;
        .page-container {
          height: 100%;
          overflow-x: auto;
          .right-page-content {
            min-width: 1250px;
            min-height: calc(100% - 32px);
            padding: 16px;
          }
        }
      }
      .no-header-main {
        height: 100%;
        .page-container {
          height: 100%;
          .right-page-content {
            height: 100%;
            padding: 0;
          }
        }
      }
    }
  }
  .outer-container ::v-deep {
    .el-submenu.is-active,
    .el-submenu.is-active > .el-submenu__title,
    .el-submenu.is-active > .el-submenu__title i:first-child,
    .el-submenu.is-active > .el-submenu__title .el-submenu__icon-arrow {
      color: $color;
    }
    .el-submenu .el-submenu__icon-arrow.el-icon-arrow-down:before {
      content: '\e790';
    }
  }
}

.menu--popover-wrap {
  border-bottom: 1px solid #ebebeb;
  padding: 4px 0 6px 0;
}
.menu--popover-wrap:first-of-type {
  padding-top: 0;
}
.menu--popover-wrap.wrap-last {
  padding-bottom: 0;
  border-bottom: none;
}
.menu--popover-item {
  font-size: 13px;
  color: $color_title;
  height: 34px;
  line-height: 34px;
  cursor: pointer;
  border-radius: 4px;
  padding: 0 8px;
  .menu--popover-item-img {
    height: 16px;
    display: inline-block;
    vertical-align: middle;
    margin-right: 5px;
  }
  .menu--popover-item-name {
    font-size: 13px;
    color: $color_title;
    display: inline-block;
    vertical-align: middle;
  }
  .menu--popover-item-icon {
    width: 16px;
    float: right;
    margin-top: 13px;
  }
  .menu--popover-item-version {
    font-size: 13px;
    float: right;
  }
  .menu--popover-item-version:after {
    display: inline-block;
    content: '';
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #f59a23;
    margin-bottom: 2px;
  }
}
.menu--popover-item:hover ::v-deep {
  background: #f5f7fa !important;
  .el-input .el-input__inner {
    border: none !important;
  }
}
</style>
